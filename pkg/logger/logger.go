package logger

import (
	"context"
	"errors"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxkey string

const CTXKEY ctxkey = "ackstream.logger"

func WithContext(ctx context.Context, l *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, CTXKEY, l)
}

func FromContext(ctx context.Context) (*zap.SugaredLogger, error) {
	l, ok := ctx.Value(CTXKEY).(*zap.SugaredLogger)
	if !ok {
		return nil, errors.New("no logger was configured")
	}

	return l, nil
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
