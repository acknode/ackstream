package app

import (
	"context"
	"time"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/xstream"
)

var TOPIC_EVENTS_PUT = "events.put"

func UsePub(ctx context.Context) func(ws, app, etype string, data interface{}) (string, error) {
	cfg := configs.FromContext(ctx)
	pub := xstream.NewPub(ctx, cfg.XStream)

	return func(ws, app, etype string, data interface{}) (string, error) {
		now := time.Now().UTC()
		e := entities.Event{
			CreationTime: now.UnixMicro(),
			Bucket:       now.Format(cfg.XStorage.BucketTemplate),
			Workspace:    ws,
			App:          app,
			Type:         etype,
		}
		e.WithId()

		return pub(TOPIC_EVENTS_PUT, &e)
	}
}

func UseSub(ctx context.Context, queue string, fn xstream.SubscribeFn) (func() error, error) {
	cfg := configs.FromContext(ctx)
	sub := xstream.NewSub(ctx, cfg.XStream)
	return sub(TOPIC_EVENTS_PUT, queue, fn)
}
