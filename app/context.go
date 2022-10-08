package app

import (
	"context"

	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/zlogger"
	"go.uber.org/zap"
)

func NewContext(ctx context.Context, logger *zap.SugaredLogger, cfg *configs.Configs) context.Context {
	ctx = configs.WithContext(ctx, cfg)
	ctx = zlogger.WithContext(ctx, logger)

	return ctx
}
