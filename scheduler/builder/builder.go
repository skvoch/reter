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
	timeStr  string
	interval time.Duration

	tickerType models.TickerType

	runner Runner
}

func (b *Builder) Seconds() *Do {
	b.interval = time.Duration(b.count) * time.Second
	b.tickerType = models.TickerInterval
	return &Do{
		builder: b,
	}
}

func (b *Builder) Minute() *Do {
	b.interval = time.Duration(b.count) * time.Minute
	b.tickerType = models.TickerInterval

	return &Do{
		builder: b,
	}
}

func (b *Builder) Interval(interval time.Duration) *Do {
	b.interval = interval
	b.tickerType = models.TickerInterval

	return &Do{
		builder: b,
	}
}

func (b *Builder) Time(time string) *Do {
	b.timeStr = time
	b.tickerType = models.TickerTime

	return &Do{
		builder: b,
	}
}

type Do struct {
	builder *Builder
}

func (d *Do) Do(ctx context.Context, name string, handler func()) error {
	task := models.Task{
		Handler:    handler,
		Interval:   d.builder.interval,
		Name:       name,
		TickerType: d.builder.tickerType,
	}

	if task.Name == "" {
		return ErrEmptyTaskName
	}

	if task.TickerType == models.TickerInterval && task.Interval == 0 {
		return ErrTaskIntervalIsZero
	}

	if task.TickerType == models.TickerTime {
		hour, minute, second, err := models.ParseTime(d.builder.timeStr)
		if err != nil {
			return fmt.Errorf("failed to parse time: %w", err)
		}
		task.Hour = hour
		task.Minute = minute
		task.Second = second
	}

	if err := d.builder.runner.Run(ctx, task); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}
