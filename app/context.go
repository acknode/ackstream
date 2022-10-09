package app

import (
	"context"

	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/acknode/ackstream/pkg/xstream"
	"github.com/acknode/ackstream/pkg/zlogger"
	"go.uber.org/zap"
)

func NewContext(ctx context.Context, logger *zap.SugaredLogger, cfg *configs.Configs) context.Context {
	ctx = configs.WithContext(ctx, cfg)
	ctx = xstorage.CfgWithContext(ctx, cfg.XStorage)

	ctx = xstream.CfgWithContext(ctx, cfg.XStream)
	ctx = zlogger.WithContext(ctx, logger)

	return ctx
}
