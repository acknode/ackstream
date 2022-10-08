package app

import (
	"context"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstream"
)

var TOPIC_EVENTS = "events"

func UsePub(ctx context.Context) (xstream.Pub, func() error) {
	cfg := configs.FromContext(ctx)
	return xstream.NewPub(ctx, cfg.XStream, TOPIC_EVENTS)
}

func UseSub(ctx context.Context, sample *entities.Event, queue string, fn xstream.SubscribeFn) (func() error, error) {
	cfg := configs.FromContext(ctx)
	sub := xstream.NewSub(ctx, cfg.XStream, TOPIC_EVENTS)
	return sub(sample, queue, fn)
}
