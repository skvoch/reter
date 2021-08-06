package scheduler

import (
	"bou.ke/monkey"
	"context"
	"github.com/aliykh/reter/scheduler/logger/zerologadapter"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/aliykh/reter/scheduler/builder"
	"github.com/aliykh/reter/scheduler/models"
	"github.com/stretchr/testify/assert"
)

func makeScheduler(taskNames ...string) *impl {
	out := &impl{
		tasks:   make(map[string]interface{}),
		tasksMx: &sync.Mutex{},
		logger:  zerologadapter.NewLogger(log.Logger),
	}

	for _, name := range taskNames {
		out.tasks[name] = struct{}{}
	}
	return out
}

func DatePtr(year int, month time.Month, day, hour, min, sec, nsec int) *time.Time {
	if year == 0 && month == 0 && day == 0 && hour == 0 && sec == 0 && nsec == 0 {
		return nil
	}

	out := time.Date(year, month, day, hour, min, sec, nsec, time.UTC)
	return &out
}

func TestEvery(t *testing.T) {
	cases := []struct {
		Name    string
		IsError bool
		Prepare func() *builder.Do
	}{
		{
			Name:    "#1",
			IsError: true,
			Prepare: func() *builder.Do {
				s := makeScheduler()
				return s.Every().Interval(0)
			},
		},
		{
			Name:    "#2",
			IsError: true,
			Prepare: func() *builder.Do {
				s := makeScheduler()
				return s.Every().Minute()
			},
		},
	}

	for _, c := range cases {
		err := c.Prepare().Do(context.Background(), "func", func() {})
		if c.IsError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestIsTimeSinceLastActionGreaterInterval(t *testing.T) {
	cases := []struct {
		Name           string
		LastActionTime *time.Time
		Now            time.Time
		Interval       time.Duration

		Expect bool
	}{
		{
			Name:           "#1",
			LastActionTime: DatePtr(2021, 01, 01, 12, 0, 0, 0),
			Now:            time.Date(2021, 01, 01, 12, 16, 0, 0, time.UTC),
			Interval:       time.Minute * 15,
			Expect:         true,
		},
		{
			Name:           "#2",
			LastActionTime: DatePtr(2021, 01, 01, 12, 0, 0, 0),
			Now:            time.Date(2021, 01, 01, 12, 13, 0, 0, time.UTC),
			Interval:       time.Minute * 15,
			Expect:         false,
		},
		{
			Name:           "#3",
			LastActionTime: nil,
			Now:            time.Date(2021, 01, 01, 12, 13, 0, 0, time.UTC),
			Interval:       time.Minute * 15,
			Expect:         true,
		},
	}

	impl := impl{}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time {
				return c.Now
			})

			result := impl.isTimeSinceLastActionGreaterInterval(c.LastActionTime, c.Interval)
			assert.Equal(t, c.Expect, result)
		})
	}

}

func TestValidateTask(t *testing.T) {

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
				Handler:  func() {},
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
