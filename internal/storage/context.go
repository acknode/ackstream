package storage

import (
	"context"
	"errors"
)

type ctxkey string

const CTXKEY_CLIENT ctxkey = "ackstream.storage.client"

func WithContext(ctx context.Context, client *Storage) context.Context {
	return context.WithValue(ctx, CTXKEY_CLIENT, client)
}

func FromContext(ctx context.Context) (*Storage, error) {
	storage, ok := ctx.Value(CTXKEY_CLIENT).(*Storage)
	if !ok {
		return nil, errors.New("no storage was configured")
	}

	return storage, nil
}
