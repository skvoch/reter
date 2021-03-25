// Package zerologadapter provides a logger that writes to a github.com/rs/zerolog.
package zerologadapter

import (
	"context"
	"github.com/skvoch/reter/scheduler"

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

func (pl *Logger) Log(ctx context.Context, level scheduler.LogLevel, msg string, data map[string]interface{}) {
	var zlevel zerolog.Level
	switch level {
	case scheduler.LogLevelNone:
		zlevel = zerolog.NoLevel
	case scheduler.LogLevelError:
		zlevel = zerolog.ErrorLevel
	case scheduler.LogLevelWarn:
		zlevel = zerolog.WarnLevel
	case scheduler.LogLevelInfo:
		zlevel = zerolog.InfoLevel
	case scheduler.LogLevelDebug:
		zlevel = zerolog.DebugLevel
	default:
		zlevel = zerolog.DebugLevel
	}

	log := pl.logger.With().Fields(data).Logger()
	log.WithLevel(zlevel).Msg(msg)
}
