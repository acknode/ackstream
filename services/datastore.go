package services

import (
	"context"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pubsub"
	"github.com/acknode/ackstream/storage"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

func NewDatastoreConsumer(ctx context.Context) (func() error, error) {
	cfg, err := configs.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	storage, err := storage.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	client, err := pubsub.FromContext[nats.Conn](ctx)
	if err != nil {
		return nil, err
	}
	jsc, err := pubsub.NewStream(client, cfg.PubSub)
	if err != nil {
		return nil, err
	}
	subscribe := pubsub.NewSub(jsc, cfg.PubSub)

	// return error will cause of re-process message
	// only return error if you got I/O error
	// @TODO: log all errors
	return subscribe("events.put", "ackstream.service.datastore", func(msg *pubsub.Message) error {
		var e event.Event
		if err := msgpack.Unmarshal(msg.Data, &e); err != nil {
			return nil
		}

		if err := storage.Put(ctx, &e); err != nil {
			return err
		}
		return nil
	})
}
