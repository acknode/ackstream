package app

import (
	"context"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstream"
)

var TOPIC_EVENTS = "events"

func UsePub(ctx context.Context) func(e *entities.Event) (string, error) {
	cfg := configs.FromContext(ctx)
	pub := xstream.NewPub(ctx, cfg.XStream)

	return func(e *entities.Event) (string, error) {
		return pub(TOPIC_EVENTS, e)
	}
}

func UseSub(ctx context.Context, sample *entities.Event, queue string, fn xstream.SubscribeFn) (func() error, error) {
	cfg := configs.FromContext(ctx)
	sub := xstream.NewSub(ctx, cfg.XStream)
	return sub(TOPIC_EVENTS, sample, queue, fn)
}
