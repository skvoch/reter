package builder

import (
	"context"
	"github.com/skvoch/reter/scheduler/models"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	m "github.com/skvoch/reter/scheduler/builder/mock"
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
					Handler:  nil,
					Interval: time.Second * 10,
					Name:     "func",
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
					Handler:  nil,
					Interval: time.Minute * 10,
					Name:     "func",
				})

				builder := New(runner, 10)
				return builder.Minute().Do(context.Background(), "func", nil)
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
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.BuildFunc(t)
			assert.Equal(t, c.Error, err)
		})
	}
}
