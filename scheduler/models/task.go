package models

import "time"

type TickerType int

const (
	TickerInterval TickerType = 0
	TickerTime     TickerType = 1
)

type Task struct {
	Name    string
	Handler func()

	TickerType           TickerType
	Interval             time.Duration
	Hour, Minute, Second int
}
