package datastore

import (
	"context"
	"errors"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/xstorage"
	"github.com/acknode/ackstream/pkg/zlogger"
)

type ctxkey string

const CTXKEY_QUEUE_NAME ctxkey = "ackstream.services.datastore.queue_name"

var ErrServiceDatastoreNoQueue = errors.New("stream queue name could not be empty")

func New(ctx context.Context) (func() error, error) {
	queue, ok := ctx.Value(CTXKEY_QUEUE_NAME).(string)
	if !ok {
		panic(ErrServiceDatastoreNoQueue)
	}

	logger := zlogger.FromContext(ctx).With("service", "datastore")
	ctx = zlogger.WithContext(ctx, logger)

	cfg := configs.FromContext(ctx)
	put := xstorage.UsePut(ctx, cfg.XStorage)

	return app.UseSub(ctx, queue, func(e *event.Event) error {
		logger.Debugw("got event", "key", e.Key())
		return put(e)
	})
}
