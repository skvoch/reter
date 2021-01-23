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
		locker: lock.NewEtcdLocker(client, lock.WithTryLockTimeout(time.Second)),
	}, nil
}

type Scheduler struct {
	tasks map[string]*Task
	opts  *Options

	locker lock.Locker
	logger Logger
}

type Task struct {
	handler  func()
	interval time.Duration
	name     string
}

func (r *Scheduler) Every(count int) *Builder {
	return newBuilder(r, count)
}

func (r *Scheduler) addTask(task *Task) error {
	_, ok := r.tasks[task.name]
	if ok {
		return ErrNotUniqueTaskName
	}
	r.tasks[task.name] = task
	return nil
}

func (r *Scheduler) Run(ctx context.Context) {
	for _, task := range r.tasks {
		go r.watcher(ctx, task)
	}
}

func (r *Scheduler) watcher(ctx context.Context, task *Task) {
	r.logger.Debug(fmt.Sprintf("run task: %s", task.name))
	ticker := time.NewTicker(task.interval)

	for {
		select {
		case <-ctx.Done():

		case <-ticker.C:
			go func() {
				var (
					err error
					l   lock.Lock
				)
				if l, err = r.locker.Acquire(task.name, int(r.opts.LockTtl.Seconds())); err != nil {
					if errors.Is(err, &lock.ErrAlreadyLocked{}) {
						r.logger.Debug(fmt.Sprintf("task %s already locked", task.name))
					}
					return
				}

				r.logger.Debug(fmt.Sprintf("task %s - lock has been locked", task.name))
				task.handler()

				if err := l.Release(); err != nil {
					r.logger.Error(fmt.Errorf("failed to realse lock: %w", err), "releasing lock")
				}
				r.logger.Debug(fmt.Sprintf("task %s - lock has been released", task.name))
			}()
		}
	}
}
