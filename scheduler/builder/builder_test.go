package builder

import (
	"context"
	"github.com/aliykh/reter/scheduler/models"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	m "github.com/aliykh/reter/scheduler/builder/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuilding(t *testing.T) {
	cases := []struct {
		Name      string
		BuildFunc func(t *testing.T) error
	}{
		{
			Name: "#1 Seconds",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)
				runner.EXPECT().Run(context.Background(), models.Task{
					Handler:    nil,
					Interval:   time.Second * 10,
					Name:       "func",
					TickerType: models.TickerInterval,
				})

				builder := New(runner, 10)
				return builder.Seconds().Do(context.Background(), "func", nil)
			},
		},
		{
			Name: "#2 Minutes",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)
				runner.EXPECT().Run(context.Background(), models.Task{
					Handler:    nil,
					Interval:   time.Minute * 10,
					Name:       "func",
					TickerType: models.TickerInterval,
				})

				builder := New(runner, 10)
				return builder.Minute().Do(context.Background(), "func", nil)
			},
		},
		{
			Name: "#3 Interval",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)
				runner.EXPECT().Run(context.Background(), models.Task{
					Handler:    nil,
					Interval:   time.Minute * 10,
					Name:       "func",
					TickerType: models.TickerInterval,
				})

				builder := New(runner, 10)
				return builder.Interval(time.Minute*10).Do(context.Background(), "func", nil)
			},
		},
		{
			Name: "#4 Time",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)
				runner.EXPECT().Run(context.Background(), models.Task{
					Handler:    nil,
					Hour:       22,
					Minute:     15,
					Second:     30,
					Name:       "func",
					TickerType: models.TickerTime,
				})

				builder := New(runner, 10)
				return builder.Time("22-15-30").Do(context.Background(), "func", nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.BuildFunc(t)
			assert.NoError(t, err)
		})
	}
}

func TestIncorrectInputs(t *testing.T) {
	cases := []struct {
		Name      string
		BuildFunc func(t *testing.T) error
		Error     error
	}{
		{
			Name: "#1 empty task name",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)

				builder := New(runner, 10)
				return builder.Seconds().Do(context.Background(), "", func() {})
			},
			Error: ErrEmptyTaskName,
		},
		{
			Name: "#2 zero task interval",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)

				builder := New(runner, 0)
				return builder.Seconds().Do(context.Background(), "task", func() {})
			},
			Error: ErrTaskIntervalIsZero,
		},
		{
			Name: "#3 invalid time format",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)

				builder := New(runner, 0)
				return builder.Time("rude-invalid-string").Do(context.Background(), "task", func() {})
			},
			Error: ErrInvalidTimeFormat,
		},
		{
			Name: "#3 invalid time range hours",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)

				builder := New(runner, 0)
				return builder.Time("24-68-00").Do(context.Background(), "task", func() {})
			},
			Error: ErrInvalidTimeFormat,
		},
		{
			Name: "#4 invalid time range minutes",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)

				builder := New(runner, 0)
				return builder.Time("25-00-00").Do(context.Background(), "task", func() {})
			},
			Error: ErrInvalidTimeFormat,
		},
		{
			Name: "#5 invalid time range seconds",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)

				builder := New(runner, 0)
				return builder.Time("23-00-88").Do(context.Background(), "task", func() {})
			},
			Error: ErrInvalidTimeFormat,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.BuildFunc(t)
			assert.ErrorIs(t, err, c.Error)
		})
	}
}
