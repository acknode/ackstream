package app

import (
	"context"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/xstream"
)

var TOPIC_EVENTS_PUT = "events.put"

func UsePub(ctx context.Context) func(e *event.Event) (string, error) {
	cfg := configs.FromContext(ctx)
	pub := xstream.NewPub(ctx, cfg.XStream)

	return func(e *event.Event) (string, error) { return pub(TOPIC_EVENTS_PUT, e) }
}

func UseSub(ctx context.Context, queue string, fn xstream.SubscribeFn) (func() error, error) {
	cfg := configs.FromContext(ctx)
	sub := xstream.NewSub(ctx, cfg.XStream)
	return sub(TOPIC_EVENTS_PUT, queue, fn)
}
