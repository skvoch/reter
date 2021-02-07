package builder

import (
	"context"
	"github.com/golang/mock/gomock"
	"testing"

	m "github.com/skvoch/reter/scheduler/builder/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	cases := []struct {
		Name      string
		BuildFunc func(t *testing.T) error
		Error     error
	}{
		{
			Name: "#1 empty handler func",
			BuildFunc: func(t *testing.T) error {
				controller := gomock.NewController(t)
				runner := m.NewMockRunner(controller)

				builder := New(runner, 10)
				return builder.Seconds().Do(context.Background(), "func", nil)
			},
			Error: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.BuildFunc(t)
			assert.Equal(t, c.Error, err)
		})
	}
}
