package scheduler

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/skvoch/reter/scheduler/models"
)

func TestValidateTask(t *testing.T) {
	makeScheduler := func(taskNames ...string) *impl {
		out := &impl{
			tasks: make(map[string]interface{}),
		}

		for _, name := range taskNames {
			out.tasks[name] = struct{}{}
		}
		return out
	}

	cases := []struct {
		Name      string
		Task      models.Task
		IsValid   bool
		Scheduler *impl
	}{
		{
			Name:      "#1 valid task",
			IsValid:   true,
			Scheduler: makeScheduler(),
			Task: models.Task{
				Interval: time.Second,
				Name:     "name",
				Handler:  func() {},
			},
		},
		{
			Name:      "#2 empty handler func",
			IsValid:   false,
			Scheduler: makeScheduler(),
			Task: models.Task{
				Interval: time.Second,
				Name:     "name",
				Handler:  nil,
			},
		},
		{
			Name:      "#3 not unique task name",
			IsValid:   false,
			Scheduler: makeScheduler("get_data"),
			Task: models.Task{
				Interval: time.Second,
				Name:     "get_data",
				Handler:  nil,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.Scheduler.validateTask(c.Task)

			if c.IsValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
