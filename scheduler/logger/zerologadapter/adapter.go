// Package zerologadapter provides a logger that writes to a github.com/rs/zerolog.
package zerologadapter

import (
	"context"
	"github.com/skvoch/reter/scheduler/logger"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

func NewLogger(logger zerolog.Logger) *Logger {
	return &Logger{
		logger: logger.With().Str("module", "reter").Logger(),
	}
}

func (pl *Logger) Log(ctx context.Context, level logger.LogLevel, msg string, data map[string]interface{}) {
	var zlevel zerolog.Level
	switch level {
	case logger.LogLevelNone:
		zlevel = zerolog.NoLevel
	case logger.LogLevelError:
		zlevel = zerolog.ErrorLevel
	case logger.LogLevelWarn:
		zlevel = zerolog.WarnLevel
	case logger.LogLevelInfo:
		zlevel = zerolog.InfoLevel
	case logger.LogLevelDebug:
		zlevel = zerolog.DebugLevel
	default:
		zlevel = zerolog.DebugLevel
	}

	log := pl.logger.With().Fields(data).Logger()
	log.WithLevel(zlevel).Msg(msg)
}
