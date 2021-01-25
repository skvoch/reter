package scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/Scalingo/go-etcd-lock/lock"
	"time"

	etcd "go.etcd.io/etcd/clientv3"
)

type EtcdOptions struct {
	Endpoints   []string
	DialTimeout time.Duration
}
type Options struct {
	Etcd    EtcdOptions
	LockTtl time.Duration
}

func New(logger Logger, opts *Options) (*Scheduler, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   opts.Etcd.Endpoints,
		DialTimeout: opts.Etcd.DialTimeout,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Scheduler{
		opts:   opts,
		tasks:  make(map[string]*Task),
		logger: logger,
		etcd:   client,
		locker: lock.NewEtcdLocker(client, lock.WithTryLockTimeout(time.Second)),
	}, nil
}

type Scheduler struct {
	tasks map[string]*Task
	opts  *Options

	locker lock.Locker
	logger Logger
	etcd   *etcd.Client
}

type Task struct {
	handler  func()
	interval time.Duration
	name     string
}

func (s *Scheduler) Every(count int) *Builder {
	return newBuilder(s, count)
}

func (s *Scheduler) runTask(ctx context.Context, task *Task) error {
	_, ok := s.tasks[task.name]
	if ok {
		return ErrNotUniqueTaskName
	}
	s.watcher(ctx, task)
	return nil
}

func (s *Scheduler) watcher(ctx context.Context, task *Task) {
	s.logger.Debug(fmt.Sprintf("run task: %s", task.name))
	ticker := time.NewTicker(task.interval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug(fmt.Sprintf("task %s - has been finished", task.name))
			return

		case <-ticker.C:
			func() {
				var (
					err error
					l   lock.Lock
				)
				lastActionTime, err := s.getLastActionTime(ctx, task.name)
				if err != nil {
					s.logger.Error(err, "getting last action time")
				}

				if lastActionTime != nil && time.Now().Sub(*lastActionTime) < task.interval {
					s.logger.Debug("too early")
					return
				}

				if l, err = s.locker.Acquire(task.name, int(s.opts.LockTtl.Seconds())); err != nil {
					if errors.Is(err, &lock.ErrAlreadyLocked{}) {
						s.logger.Debug(fmt.Sprintf("task %s already locked", task.name))
						return
					}
					s.logger.Error(err, "trying to acquire")
					return
				}

				s.logger.Debug(fmt.Sprintf("task %s - lock has been locked", task.name))
				task.handler()

				if err := s.setLastActionTime(ctx, task.name, time.Now()); err != nil {
					s.logger.Error(err, "setting last action time")
					return
				}

				if err := l.Release(); err != nil {
					s.logger.Error(fmt.Errorf("failed to realse lock: %w", err), "releasing lock")
					return
				}
				s.logger.Debug(fmt.Sprintf("task %s - lock has been released", task.name))
			}()
		}
	}
}

func (s *Scheduler) getLastActionTime(ctx context.Context, taskName string) (*time.Time, error) {
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

func (s *Scheduler) setLastActionTime(ctx context.Context, taskName string, t time.Time) error {
	if _, err := s.etcd.Put(ctx, taskName, t.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("failed to set last action time: %w", err)
	}
	return nil
}
