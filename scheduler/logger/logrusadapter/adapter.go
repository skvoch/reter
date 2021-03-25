// Package logrusadapter provides a logger that writes to a github.com/sirupsen/logrus.Logger
// log.
package logrusadapter

import (
	"context"
	"github.com/skvoch/reter/scheduler"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	l logrus.FieldLogger
}

func NewLogger(l logrus.FieldLogger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Log(ctx context.Context, level scheduler.LogLevel, msg string, data map[string]interface{}) {
	var logger logrus.FieldLogger
	if data != nil {
		logger = l.l.WithFields(data)
	} else {
		logger = l.l
	}

	switch level {
	case scheduler.LogLevelTrace:
		logger.WithField("RETER_LOG_LEVEL", level).Debug(msg)
	case scheduler.LogLevelDebug:
		logger.Debug(msg)
	case scheduler.LogLevelInfo:
		logger.Info(msg)
	case scheduler.LogLevelWarn:
		logger.Warn(msg)
	case scheduler.LogLevelError:
		logger.Error(msg)
	default:
		logger.WithField("INVALID_RETER_LOG_LEVEL", level).Error(msg)
	}
}
