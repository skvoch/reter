// Package logrusadapter provides a logger that writes to a github.com/sirupsen/logrus.Logger
// log.
package logrusadapter

import (
	"context"
	"github.com/aliykh/reter/scheduler/logger"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	l logrus.FieldLogger
}

func NewLogger(l logrus.FieldLogger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Log(ctx context.Context, level logger.LogLevel, msg string, data map[string]interface{}) {
	var log logrus.FieldLogger
	if data != nil {
		log = l.l.WithFields(data)
	} else {
		log = l.l
	}

	switch level {
	case logger.LogLevelTrace:
		log.WithField("RETER_LOG_LEVEL", level).Debug(msg)
	case logger.LogLevelDebug:
		log.Debug(msg)
	case logger.LogLevelInfo:
		log.Info(msg)
	case logger.LogLevelWarn:
		log.Warn(msg)
	case logger.LogLevelError:
		log.Error(msg)
	default:
		log.WithField("INVALID_RETER_LOG_LEVEL", level).Error(msg)
	}
}
