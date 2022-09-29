package app

import (
	"context"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/stream"
)

var TOPIC_EVENTS_PUT = "events.put"

func NewPub(ctx context.Context) func(e *event.Event) (string, error) {
	cfg := configs.FromContext(ctx)
	pub := stream.NewPub(ctx, cfg.Stream)

	return func(e *event.Event) (string, error) {
		return pub(TOPIC_EVENTS_PUT, e)
	}
}

func NewSub(ctx context.Context, queue string, fn stream.SubscribeFn) (func() error, error) {
	cfg := configs.FromContext(ctx)
	sub := stream.NewSub(ctx, cfg.Stream)
	return sub(TOPIC_EVENTS_PUT, queue, fn)
}
