package logger

import "github.com/rs/zerolog"

func Zerolog(logger zerolog.Logger) Logger {
	return &zerologImpl{
		logger: logger,
	}
}

type zerologImpl struct {
	logger zerolog.Logger
}

func (z *zerologImpl) Info(msg string) {
	z.logger.Info().Msg(msg)
}

func (z *zerologImpl) Debug(msg string) {
	z.logger.Debug().Msg(msg)
}

func (z *zerologImpl) Error(err error, msg string) {
	z.logger.Error().Err(err).Msg(msg)
}
