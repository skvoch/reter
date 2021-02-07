package models

import "time"

type Task struct {
	Handler  func()
	Interval time.Duration
	Name     string
}
