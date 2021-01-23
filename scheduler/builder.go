package scheduler

import (
	"fmt"
	"time"
)

func newBuilder(instance *Scheduler, count int) *Builder {
	return &Builder{
		count:         count,
		reterInstance: instance,
	}
}

type Builder struct {
	count    int
	interval time.Duration

	reterInstance *Scheduler
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

func (d *Do) Do(name string, handler func()) error {
	if handler == nil {
		return ErrEmptyHandler
	}

	if name == "" {
		return ErrEmptyTask
	}

	task := &Task{
		handler:  handler,
		interval: d.builder.interval,
		name:     name,
	}

	if err := d.builder.reterInstance.addTask(task); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}
