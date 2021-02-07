package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Scalingo/go-etcd-lock/lock"
	"github.com/skvoch/reter/scheduler/builder"
	"github.com/skvoch/reter/scheduler/models"

	etcd "go.etcd.io/etcd/clientv3"
)

var (
	ErrNotUniqueTaskName = errors.New("not unique task name")
	ErrNilHandler        = errors.New("handler func is nil")
)

type EtcdOptions struct {
	Endpoints   []string
	DialTimeout time.Duration
}

type Options struct {
	Etcd    EtcdOptions
	LockTTL time.Duration
}

type Scheduler interface {
	Every(count uint) *builder.Builder
}

func New(logger Logger, opts *Options) (Scheduler, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   opts.Etcd.Endpoints,
		DialTimeout: opts.Etcd.DialTimeout,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &impl{
		opts:   opts,
		tasks:  make(map[string]interface{}),
		logger: logger,
		etcd:   client,
		locker: lock.NewEtcdLocker(client, lock.WithTryLockTimeout(time.Second)),
	}, nil
}

type impl struct {
	tasks map[string]interface{}
	opts  *Options

	locker lock.Locker
	logger Logger
	etcd   *etcd.Client
}

func (s *impl) Every(count uint) *builder.Builder {
	return builder.New(s, count)
}

func (s *impl) Run(ctx context.Context, task models.Task) error {
	if err := s.validateTask(task); err != nil {
		return fmt.Errorf("failed to validate task: %w", err)
	}

	s.tasks[task.Name] = struct{}{}
	s.watcher(ctx, task)

	return nil
}

func (s *impl) validateTask(task models.Task) error {
	if task.Handler == nil {
		return ErrNilHandler
	}

	_, ok := s.tasks[task.Name]
	if ok {
		return ErrNotUniqueTaskName
	}
	return nil
}

func (s *impl) watcher(ctx context.Context, task models.Task) {
	s.logger.Debug(fmt.Sprintf("run task: %s", task.Name))
	ticker := time.NewTicker(task.Interval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug(fmt.Sprintf("task %s - has been finished", task.Name))
			return

		case <-ticker.C:
			func() {
				var (
					err error
					l   lock.Lock
				)
				lastActionTime, err := s.getLastActionTime(ctx, task.Name)
				if err != nil {
					s.logger.Error(err, "getting last action time")
					return
				}

				if lastActionTime != nil && time.Since(*lastActionTime) < task.Interval {
					s.logger.Debug("too early")
					return
				}

				if l, err = s.locker.Acquire(task.Name, int(s.opts.LockTTL.Seconds())); err != nil {
					if errors.Is(err, &lock.ErrAlreadyLocked{}) {
						s.logger.Debug(fmt.Sprintf("task %s already locked", task.Name))
						return
					}
					s.logger.Error(err, "trying to acquire")
					return
				}

				s.logger.Debug(fmt.Sprintf("task %s - lock has been locked", task.Name))
				task.Handler()

				if err := s.setLastActionTime(ctx, task.Name, time.Now()); err != nil {
					s.logger.Error(err, "setting last action time")
					return
				}

				if err := l.Release(); err != nil {
					s.logger.Error(fmt.Errorf("failed to release lock: %w", err), "releasing lock")
					return
				}
				s.logger.Debug(fmt.Sprintf("task %s - lock has been released", task.Name))
			}()
		}
	}
}

func (s *impl) getLastActionTime(ctx context.Context, taskName string) (*time.Time, error) {
	res, err := s.etcd.Get(ctx, taskName)
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

func (s *impl) setLastActionTime(ctx context.Context, taskName string, t time.Time) error {
	if _, err := s.etcd.Put(ctx, taskName, t.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("failed to set last action time: %w", err)
	}
	return nil
}
