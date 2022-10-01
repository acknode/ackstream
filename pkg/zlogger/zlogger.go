package zlogger

import (
	"context"
	"errors"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func New(debug bool) *zap.SugaredLogger {
	ws := zapcore.Lock(os.Stdout)

	priority := withEnableLevel(debug)
	encoder := withEncoder(debug)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, ws, priority),
	)
	logger := zap.New(core)
	defer logger.Sync()

	return logger.Sugar()
}

func withEnableLevel(debug bool) zap.LevelEnablerFunc {
	return zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if debug {
			return lvl >= zapcore.DebugLevel
		}
		return lvl >= zapcore.InfoLevel
	})
}

func withEncoder(debug bool) zapcore.Encoder {
	if debug {
		return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}
