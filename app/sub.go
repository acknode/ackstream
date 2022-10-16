package app

import (
	"context"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xstream"
)

type Sub func(sample *entities.Event, queue string, fn xstream.SubscribeFn) error

func NewSub(ctx context.Context) (Sub, error) {
	sub, err := xstream.NewSub(ctx)
	if err != nil {
		return nil, err
	}

	return func(sample *entities.Event, queue string, fn xstream.SubscribeFn) error {
		return sub(sample, queue, fn)
	}, nil
}
