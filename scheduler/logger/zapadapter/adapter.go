// Package zapadapter provides a logger that writes to a go.uber.org/zap.Logger.
package zapadapter

import (
	"context"
	"github.com/aliykh/reter/scheduler/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{logger: logger.WithOptions(zap.AddCallerSkip(1))}
}

func (pl *Logger) Log(ctx context.Context, level logger.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zapcore.Field, len(data))
	i := 0
	for k, v := range data {
		fields[i] = zap.Reflect(k, v)
		i++
	}

	switch level {
	case logger.LogLevelTrace:
		pl.logger.Debug(msg, append(fields, zap.Stringer("RETER_LOG_LEVEL", level))...)
	case logger.LogLevelDebug:
		pl.logger.Debug(msg, fields...)
	case logger.LogLevelInfo:
		pl.logger.Info(msg, fields...)
	case logger.LogLevelWarn:
		pl.logger.Warn(msg, fields...)
	case logger.LogLevelError:
		pl.logger.Error(msg, fields...)
	default:
		pl.logger.Error(msg, append(fields, zap.Stringer("INVALID_RETER_LOG_LEVEL", level))...)
	}
}
