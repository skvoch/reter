package scheduler

import "github.com/rs/zerolog"

type Logger interface {
	Info(msg string)
	Debug(msg string)
	Error(err error, msg string)
}

func Zerolog(logger zerolog.Logger) Logger {
	return &ZerologImpl{
		logger: logger,
	}
}

type ZerologImpl struct {
	logger zerolog.Logger
}

func (z *ZerologImpl) Info(msg string) {
	z.logger.Info().Msg(msg)
}

func (z *ZerologImpl) Debug(msg string) {
	z.logger.Debug().Msg(msg)
}

func (z *ZerologImpl) Error(err error, msg string) {
	z.logger.Error().Err(err).Msg(msg)
}
