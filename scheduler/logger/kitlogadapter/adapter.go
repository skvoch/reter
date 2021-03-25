package kitlogadapter

import (
	"context"
	"github.com/skvoch/reter/scheduler"

	"github.com/go-kit/kit/log"
	kitlevel "github.com/go-kit/kit/log/level"
)

type Logger struct {
	l log.Logger
}

func NewLogger(l log.Logger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Log(ctx context.Context, level scheduler.LogLevel, msg string, data map[string]interface{}) {
	logger := l.l
	for k, v := range data {
		logger = log.With(logger, k, v)
	}

	switch level {
	case scheduler.LogLevelTrace:
		logger.Log("RETER_LOG_LEVEL", level, "msg", msg)
	case scheduler.LogLevelDebug:
		kitlevel.Debug(logger).Log("msg", msg)
	case scheduler.LogLevelInfo:
		kitlevel.Info(logger).Log("msg", msg)
	case scheduler.LogLevelWarn:
		kitlevel.Warn(logger).Log("msg", msg)
	case scheduler.LogLevelError:
		kitlevel.Error(logger).Log("msg", msg)
	default:
		logger.Log("INVALID_RETER_LOG_LEVEL", level, "error", msg)
	}
}
