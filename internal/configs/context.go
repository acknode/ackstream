package configs

import (
	"context"
	"errors"
	"github.com/acknode/ackstream/pkg/xrpc"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/acknode/ackstream/pkg/xstream"
)

type ctxkey string

const CTXKEY ctxkey = "ackstream.configs"

func WithContext(ctx context.Context, cfg *Configs) context.Context {
	ctx = xstream.CfgWithContext(ctx, cfg.XStream)
	ctx = xstorage.CfgWithContext(ctx, cfg.XStorage)
	ctx = xrpc.CfgWithContext(ctx, cfg.XRPC)
	return context.WithValue(ctx, CTXKEY, cfg)
}

func FromContext(ctx context.Context) *Configs {
	configs, ok := ctx.Value(CTXKEY).(*Configs)
	if !ok {
		panic(errors.New("no configs was configured"))
	}

	return configs
}
