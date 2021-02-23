package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/skvoch/go-etcd-lock/v5/lock"
	"github.com/skvoch/reter/scheduler/builder"
	"github.com/skvoch/reter/scheduler/models"

	etcd "go.etcd.io/etcd/v3/clientv3"
)

var (
	ErrNotUniqueTaskName = errors.New("not unique task name")
	ErrNilHandler        = errors.New("handler func is nil")
)

type EtcdOptions struct {
	Endpoints []string
}

type Options struct {
	Etcd    EtcdOptions
	LockTTL time.Duration
	Timeout time.Duration
}

type Scheduler interface {
	Every(count uint) *builder.Builder
}

func New(logger Logger, opts *Options) (Scheduler, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints: opts.Etcd.Endpoints,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &impl{
		opts:   opts,
		tasks:  make(map[string]interface{}),
		logger: logger,
		etcd:   client,
		locker: lock.NewEtcdLocker(client, lock.WithMaxTryLockTimeout(opts.Timeout)),
	}, nil
}

type impl struct {
	tasks map[string]interface{}
	opts  *Options

	locker lock.Locker
	logger Logger
	etcd   *etcd.Client
}

func (i *impl) Every(count uint) *builder.Builder {
	return builder.New(i, count)
}

func (i *impl) Run(ctx context.Context, task models.Task) error {
	if err := i.validateTask(task); err != nil {
		return fmt.Errorf("failed to validate task: %w", err)
	}

	i.tasks[task.Name] = struct{}{}
	i.watcher(ctx, task)

	return nil
}

func (i *impl) validateTask(task models.Task) error {
	if task.Handler == nil {
		return ErrNilHandler
	}

	_, ok := i.tasks[task.Name]
	if ok {
		return ErrNotUniqueTaskName
	}
	return nil
}

func (i *impl) watcher(ctx context.Context, task models.Task) {
	i.logger.Debug(fmt.Sprintf("run task: %s", task.Name))
	ticker := time.NewTicker(task.Interval)

	for {
		select {
		case <-ctx.Done():
			i.logger.Debug(fmt.Sprintf("task %s - has been finished", task.Name))
			return

		case <-ticker.C:
			if err := i.handler(ctx, task); err != nil {
				i.logger.Error(err, "trying to run handler function")
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
		i.logger.Debug("time since last action less than interval")
		return nil
	}

	if l, err = i.locker.Acquire(i.contextWithTimeout(ctx), task.Name, int(i.opts.LockTTL.Seconds())); err != nil {
		if errors.Is(err, &lock.ErrAlreadyLocked{}) {
			i.logger.Debug(fmt.Sprintf("task %s already locked", task.Name))
			return nil
		}
		return fmt.Errorf("failed to acquire locker: %w", err)
	}
	i.logger.Debug(fmt.Sprintf("task %s - locker has been locked", task.Name))

	task.Handler()

	if err := i.setLastActionTime(i.contextWithTimeout(ctx), task.Name, time.Now()); err != nil {
		return fmt.Errorf("failed to set last action time: %w", err)
	}

	if err := l.Release(); err != nil {
		return fmt.Errorf("failed to release locker: %w", err)
	}
	i.logger.Debug(fmt.Sprintf("task %s - lock has been released", task.Name))
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
		return nil, fmt.Errorf("failed to get last action time: %w", err)
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
