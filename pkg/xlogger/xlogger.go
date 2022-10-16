package xlogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func New(debug bool) *zap.SugaredLogger {
	ws := zapcore.Lock(os.Stdout)

	priority := withEnableLevel(debug)
	encoder := withEncoder(debug)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, ws, priority),
	)
	logger := zap.New(core)
	return logger.Sugar()
}

func withEnableLevel(debug bool) zap.LevelEnablerFunc {
	return func(lvl zapcore.Level) bool {
		if debug {
			return lvl >= zapcore.DebugLevel
		}
		return lvl >= zapcore.InfoLevel
	}
}

func withEncoder(debug bool) zapcore.Encoder {
	if debug {
		return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}
