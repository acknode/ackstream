package app

import (
	"context"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/acknode/ackstream/pkg/xstream"
)

type Pub func(event *entities.Event) (*string, error)

func NewPub(ctx context.Context) (Pub, error) {
	logger := xlogger.FromContext(ctx).With("app", "publisher")

	put, err := xstorage.NewPut(ctx)
	if err != nil {
		return nil, err
	}

	pub, err := xstream.NewPub(ctx)
	if err != nil {
		return nil, err
	}

	return func(event *entities.Event) (*string, error) {
		flogger := logger.With("event_key", event.Key())

		if err := put(event); err != nil {
			return nil, err
		}
		flogger.Debugw("put event")

		pubkey, err := pub(event)
		flogger.Debugw("published", "pubkey", pubkey)

		return pubkey, err
	}, nil
}
