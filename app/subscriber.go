package app

import (
	"context"

	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/pubsub"
)

func NewSubscriber(ctx context.Context) (pubsub.Sub, error) {
	cfg, err := configs.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	client, err := pubsub.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	jsc, err := pubsub.NewStream(client, cfg.PubSub)
	if err != nil {
		return nil, err
	}

	return pubsub.NewSub(jsc, cfg.PubSub), nil
}
