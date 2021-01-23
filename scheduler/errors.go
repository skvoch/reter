package scheduler

import "errors"

var (
	ErrEmptyHandler      = errors.New("empty handler func")
	ErrEmptyTask         = errors.New("empty task name")
	ErrNotUniqueTaskName = errors.New("not unique task name")
)
