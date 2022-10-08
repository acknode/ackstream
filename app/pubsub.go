package app

import (
	"context"

	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstream"
)

var TOPIC_EVENTS = "events"

func UsePub(ctx context.Context) (xstream.Pub, func() error) {
	cfg := configs.FromContext(ctx)
	return xstream.NewPub(ctx, cfg.XStream, TOPIC_EVENTS)
}

func UseSub(ctx context.Context) xstream.Sub {
	cfg := configs.FromContext(ctx)
	return xstream.NewSub(ctx, cfg.XStream, TOPIC_EVENTS)
}
