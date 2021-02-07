package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/skvoch/reter/scheduler/models"
	"time"
)

var (
	ErrEmptyHandler  = errors.New("empty handler func")
	ErrEmptyTaskName = errors.New("empty task name")
)

type Runner interface {
	Run(ctx context.Context, task *models.Task) error
}

func New(runner Runner, count int) *Builder {
	return &Builder{
		count:  count,
		runner: runner,
	}
}

type Builder struct {
	count    int
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

type Do struct {
	builder *Builder
}

func (d *Do) Do(ctx context.Context, name string, handler func()) error {
	if handler == nil {
		return ErrEmptyHandler
	}

	if name == "" {
		return ErrEmptyTaskName
	}

	task := &models.Task{
		Handler:  handler,
		Interval: d.builder.interval,
		Name:     name,
	}

	if err := d.builder.runner.Run(ctx, task); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}
