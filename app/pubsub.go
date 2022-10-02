package app

import (
	"context"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/xstream"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/utils"
)

var TOPIC_EVENTS_PUT = "events.put"

func UsePub(ctx context.Context) func(ws, app, etype string, data interface{}) (string, error) {
	cfg := configs.FromContext(ctx)
	pub := xstream.NewPub(ctx, cfg.XStream)

	return func(ws, app, etype string, data interface{}) (string, error) {
		bucket, ts := utils.NewBucket(cfg.XStorage.BucketTemplate)
		e := entities.Event{
			Bucket:    bucket,
			Workspace: ws,
			App:       app,
			Type:      etype,

			CreationTime: ts,
			Data:         data,
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
