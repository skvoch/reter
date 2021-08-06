package kitlogadapter

import (
	"context"
	"github.com/aliykh/reter/scheduler/logger"

	"github.com/go-kit/kit/log"
	kitlevel "github.com/go-kit/kit/log/level"
)

type Logger struct {
	l log.Logger
}

func NewLogger(l log.Logger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Log(ctx context.Context, level logger.LogLevel, msg string, data map[string]interface{}) {
	lg := l.l
	for k, v := range data {
		lg = log.With(lg, k, v)
	}

	switch level {
	case logger.LogLevelTrace:
		lg.Log("RETER_LOG_LEVEL", level, "msg", msg)
	case logger.LogLevelDebug:
		kitlevel.Debug(lg).Log("msg", msg)
	case logger.LogLevelInfo:
		kitlevel.Info(lg).Log("msg", msg)
	case logger.LogLevelWarn:
		kitlevel.Warn(lg).Log("msg", msg)
	case logger.LogLevelError:
		kitlevel.Error(lg).Log("msg", msg)
	default:
		lg.Log("INVALID_RETER_LOG_LEVEL", level, "error", msg)
	}
}
