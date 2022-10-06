package datastore

import (
	"context"
	"errors"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/xstorage"
	"github.com/acknode/ackstream/internal/xstream"
	"github.com/acknode/ackstream/pkg/configs"
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

	return app.UseSub(ctx, queue, UseHandler(ctx))
}

func UseHandler(ctx context.Context) xstream.SubscribeFn {
	logger := zlogger.FromContext(ctx).With("service", "datastore")
	ctx = zlogger.WithContext(ctx, logger)

	cfg := configs.FromContext(ctx)
	put := xstorage.UsePut(ctx, cfg.XStorage)

	return func(e *entities.Event) error {
		logger.Debugw("got entities", "key", e.Key())
		return put(e)
	}
}
