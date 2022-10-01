package app

import (
	"context"

	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/xstream"
	"github.com/acknode/ackstream/pkg/zlogger"
	"go.uber.org/zap"
)

func NewContext(ctx context.Context, logger *zap.SugaredLogger, cfg *configs.Configs) (context.Context, func()) {
	ctx = configs.WithContext(ctx, cfg)
	ctx = zlogger.WithContext(ctx, logger)

	stream, conn := xstream.New(ctx, cfg.XStream)
	ctx = xstream.WithContext(ctx, stream, conn)

	return ctx, func() { conn.Drain() }
}
