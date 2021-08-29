package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/skvoch/reter/scheduler/logger"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/skvoch/go-etcd-lock/v5/lock"
	"github.com/skvoch/reter/scheduler/builder"
	"github.com/skvoch/reter/scheduler/models"
)

var (
	ErrNotUniqueTaskName = errors.New("not unique task name")
	ErrNilHandler        = errors.New("handler func is nil")
)

type EtcdOptions struct {
	Endpoints   []string
	LogWarnings bool
}

type Options struct {
	Etcd    EtcdOptions
	LockTTL time.Duration
	Timeout time.Duration
}

type Scheduler interface {
	Every(count ...uint) *builder.Builder
}

func New(logger logger.Logger, opts *Options) (Scheduler, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level.SetLevel(zap.ErrorLevel)
	if opts.Etcd.LogWarnings {
		zapConfig.Level.SetLevel(zap.WarnLevel)
	}

	client, err := etcd.New(etcd.Config{
		Endpoints: opts.Etcd.Endpoints,
		LogConfig: &zapConfig,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &impl{
		opts:    opts,
		tasks:   make(map[string]interface{}),
		logger:  logger,
		tasksMx: &sync.Mutex{},
		etcd:    client,
		locker:  lock.NewEtcdLocker(client, lock.WithMaxTryLockTimeout(opts.Timeout)),
	}, nil
}

type impl struct {
	tasksMx *sync.Mutex
	tasks   map[string]interface{}
	opts    *Options

	locker lock.Locker
	logger logger.Logger
	etcd   *etcd.Client
}

func (i *impl) Every(inputCount ...uint) *builder.Builder {
	var count uint

	if len(inputCount) != 0 {
		count = inputCount[0]
	}

	return builder.New(i, count)
}

func (i *impl) Run(ctx context.Context, task models.Task) error {
	if err := i.validateTask(task); err != nil {
		return fmt.Errorf("failed to validate task: %w", err)
	}

	i.setTask(task.Name)

	switch task.TickerType {
	case models.TickerInterval:
		i.watcherInterval(ctx, task)
	case models.TickerTime:
		i.watcherTime(ctx, task)
	default:
		return fmt.Errorf("failed to run task: unknown ticker type %v", task.TickerType)
	}

	return nil
}

func (i *impl) setTask(name string) {
	i.tasksMx.Lock()
	defer i.tasksMx.Unlock()

	i.tasks[name] = struct{}{}
}

func (i *impl) validateTask(task models.Task) error {
	if task.Handler == nil {
		return ErrNilHandler
	}

	i.tasksMx.Lock()
	defer i.tasksMx.Unlock()

	_, ok := i.tasks[task.Name]
	if ok {
		return ErrNotUniqueTaskName
	}
	return nil
}

func (i *impl) watcherInterval(ctx context.Context, task models.Task) {
	i.logger.Log(ctx, logger.LogLevelInfo, "running task", map[string]interface{}{"task_name": task.Name})
	ticker := time.NewTicker(task.Interval)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			i.logger.Log(ctx, logger.LogLevelInfo, "task has been finished", map[string]interface{}{"task_name": task.Name})
			return

		case <-ticker.C:
			if err := i.handler(ctx, task); err != nil {
				i.logger.Log(ctx, logger.LogLevelError, "trying to run handler function", map[string]interface{}{"error": err})
			}
		}
	}
}

func (i *impl) watcherTime(ctx context.Context, task models.Task) {
	i.logger.Log(ctx, logger.LogLevelInfo, "running task", map[string]interface{}{"task_name": task.Name})

MainLoop:
	for {
		now := time.Now()
		target := time.Date(now.Year(), now.Month(), now.Day(), task.Hour, task.Minute, task.Second, 0, now.Location())
		diff := target.Sub(now)

		if diff < 0 {
			diff = diff + (time.Hour * 24)
		}
		timer := time.NewTimer(diff)

		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				i.logger.Log(ctx, logger.LogLevelInfo, "task has been finished", map[string]interface{}{"task_name": task.Name})
				return

			case <-timer.C:
				for {
					select {
					case <-ctx.Done():
						timer.Stop()
						i.logger.Log(ctx, logger.LogLevelInfo, "task has been finished", map[string]interface{}{"task_name": task.Name})
						return

					default:
						if err := i.handler(ctx, task); err != nil {
							i.logger.Log(ctx, logger.LogLevelError, "trying to run handler function", map[string]interface{}{"error": err})
							time.Sleep(time.Second * 3)
							continue
						}
						timer.Stop()
						goto MainLoop
					}
				}
			}
		}
	}
}

func (i *impl) handler(ctx context.Context, task models.Task) error {
	var (
		err error
		l   lock.Lock
	)
	lastActionTime, err := i.getLastActionTime(i.contextWithTimeout(ctx), task.Name)
	if err != nil {
		return fmt.Errorf("failed to get last action time: %w", err)
	}

	if !i.isTimeSinceLastActionGreaterInterval(lastActionTime, task.Interval) {
		i.logger.Log(ctx, logger.LogLevelDebug, "time since last action less than interval", map[string]interface{}{"task_name": task.Name})
		return nil
	}

	if l, err = i.locker.Acquire(i.contextWithTimeout(ctx), task.Name, int(i.opts.LockTTL.Seconds())); err != nil {
		if errors.Is(err, &lock.ErrAlreadyLocked{}) {
			i.logger.Log(ctx, logger.LogLevelDebug, "task already locked", map[string]interface{}{"task_name": task.Name})
			return nil
		}
		return fmt.Errorf("failed to acquire locker: %w", err)
	}

	i.logger.Log(ctx, logger.LogLevelDebug, "locker has been locked", map[string]interface{}{"task_name": task.Name})

	task.Handler()

	if err := i.setLastActionTime(i.contextWithTimeout(ctx), task.Name, time.Now()); err != nil {
		return fmt.Errorf("failed to set last action time: %w", err)
	}

	if err := l.Release(); err != nil {
		return fmt.Errorf("failed to release locker: %w", err)
	}

	i.logger.Log(ctx, logger.LogLevelDebug, "locker has been released", map[string]interface{}{"task_name": task.Name})
	return nil
}

func (i *impl) contextWithTimeout(ctx context.Context) context.Context {
	out, _ := context.WithTimeout(ctx, i.opts.Timeout)
	return out
}

func (i *impl) isTimeSinceLastActionGreaterInterval(lastActionTime *time.Time, interval time.Duration) bool {
	if lastActionTime == nil {
		return true
	}

	return time.Since(*lastActionTime) > interval
}

func (i *impl) getLastActionTime(ctx context.Context, taskName string) (*time.Time, error) {
	res, err := i.etcd.Get(ctx, taskName)
	if err != nil {
		return nil, err
	}
	if len(res.Kvs) == 0 {
		return nil, nil
	}
	out, err := time.Parse(time.RFC3339, string(res.Kvs[0].Value))
	if err != nil {
		return nil, fmt.Errorf("failed to parse last action time: %w", err)
	}
	return &out, nil
}

func (i *impl) setLastActionTime(ctx context.Context, taskName string, t time.Time) error {
	if _, err := i.etcd.Put(ctx, taskName, t.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("failed to set last action time: %w", err)
	}
	return nil
}
