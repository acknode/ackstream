package datastore

import (
	"context"
	"errors"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/logger"
	"github.com/acknode/ackstream/internal/storage"
)

type ctxkey string

const CTXKEY_QUEUE_NAME ctxkey = "ackstream.services.datastore.queue_name"

var ErrServiceDatastoreNoQueue = errors.New("stream queue name could not be empty")

func New(ctx context.Context) (func() error, error) {
	queue, ok := ctx.Value(CTXKEY_QUEUE_NAME).(string)
	if !ok {
		panic(ErrServiceDatastoreNoQueue)
	}

	s, err := storage.FromContext(ctx)
	if err != nil {
		panic(err)
	}

	l := logger.FromContext(ctx).With("service", "datastore")
	ctx = logger.WithContext(ctx, l)
	return app.NewSub(ctx, queue, func(e *event.Event) error {
		l.Debugw("got event", "key", e.Key())
		return s.Put(ctx, e)
	})
}
