package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/skvoch/reter/scheduler/models"
	"time"
)

var (
	ErrEmptyTaskName      = errors.New("task name is nil")
	ErrTaskIntervalIsZero = errors.New("task interval is zero")
)

type Runner interface {
	Run(ctx context.Context, task models.Task) error
}

func New(runner Runner, count uint) *Builder {
	return &Builder{
		count:  count,
		runner: runner,
	}
}

type Builder struct {
	count    uint
	interval time.Duration

	runner Runner
}

func (b *Builder) Seconds() *Do {
	b.interval = time.Duration(b.count) * time.Second
	return &Do{
		builder: b,
	}
}

func (b *Builder) Minute() *Do {
	b.interval = time.Duration(b.count) * time.Minute
	return &Do{
		builder: b,
	}
}

func (b *Builder) Interval(interval time.Duration) *Do {
	b.interval = interval
	return &Do{
		builder: b,
	}
}

type Do struct {
	builder *Builder
}

func (d *Do) Do(ctx context.Context, name string, handler func()) error {
	task := models.Task{
		Handler:  handler,
		Interval: d.builder.interval,
		Name:     name,
	}

	if task.Name == "" {
		return ErrEmptyTaskName
	}

	if task.Interval == 0 {
		return ErrTaskIntervalIsZero
	}

	if err := d.builder.runner.Run(ctx, task); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}
