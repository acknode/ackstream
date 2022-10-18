package configs

import (
	"context"
	"errors"
)

type ctxkey string

const CTXKEY ctxkey = "ackstream.services.events.configs"

func WithContext(ctx context.Context, cfg *Configs) context.Context {
	return context.WithValue(ctx, CTXKEY, cfg)
}

func FromContext(ctx context.Context) *Configs {
	configs, ok := ctx.Value(CTXKEY).(*Configs)
	if !ok {
		panic(errors.New("no configs was configured"))
	}

	return configs
}
