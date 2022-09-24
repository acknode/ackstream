package configs

import (
	"context"
	"errors"

	"github.com/acknode/ackstream/pubsub"
	"github.com/acknode/ackstream/storage"
)

type Configs struct {
	PubSub  *pubsub.Configs
	Storage *storage.Configs
}

type ctxkey string

const CTXKEY ctxkey = "ackstream.configs"

func FromContext(ctx context.Context) (*Configs, error) {
	configs, ok := ctx.Value(CTXKEY).(*Configs)
	if !ok {
		return nil, errors.New("no configs was configured")
	}

	return configs, nil
}
