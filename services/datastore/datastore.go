package datastore

import (
	"context"
	"errors"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/storage"
	"github.com/acknode/ackstream/pkg/pubsub"
	"github.com/vmihailenco/msgpack/v5"
)

type ctxkey string

const CTXKEY_QUEUE_NAME ctxkey = "ackstream.services.datastore.queue_name"

var ErrServiceDatastoreNoQueue = errors.New("pubsub queue name could not be empty")

func New(ctx context.Context) (func() error, error) {
	sub, err := app.NewSubscriber(ctx)
	if err != nil {
		return nil, err
	}

	s, err := storage.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	queue, ok := ctx.Value(CTXKEY_QUEUE_NAME).(string)
	if !ok {
		return nil, ErrServiceDatastoreNoQueue
	}

	return sub(event.TOPIC_EVENT_PUT, queue, func(msg *pubsub.Message) error {
		var e event.Event
		if err := msgpack.Unmarshal(msg.Data, &e); err != nil {
			return nil
		}

		return s.Put(ctx, &e)
	})
}
