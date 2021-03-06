package logger

import (
	"go.uber.org/zap"
)

func Zap(logger *zap.Logger) *zapImpl {
	return &zapImpl{
		logger: logger,
	}
}

type zapImpl struct {
	logger *zap.Logger
}

func (z *zapImpl) Info(msg string) {
	z.logger.Info(msg)
}

func (z *zapImpl) Debug(msg string) {
	z.logger.Debug(msg)
}

func (z *zapImpl) Error(err error, msg string) {
	z.logger.Error(msg, zap.Error(err))
}
