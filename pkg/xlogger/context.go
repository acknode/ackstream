package xlogger

import (
	"context"
	"errors"
	"go.uber.org/zap"
)

type ctxkey string

const CTXKEY ctxkey = "ack.zlogger"

func WithContext(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, CTXKEY, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	logger, ok := ctx.Value(CTXKEY).(*zap.SugaredLogger)
	if !ok {
		panic(errors.New("no logger was configured"))
	}

	return logger
}
