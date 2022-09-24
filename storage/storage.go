package storage

import (
	"context"
	"errors"

	"github.com/acknode/ackstream/event"
)

type Storage interface {
	Start() error
	Stop() error

	Put(ctx context.Context, e *event.Event) error
	Get(ctx context.Context, bucket, workspace, app, etype string, id string) (*event.Event, error)
	Scan(ctx context.Context, bucket, workspace, app, etype string, size int, page []byte) ([]event.Event, []byte, []error)
}

type Configs struct {
	Hosts    []string `json:"hosts" mapstructure:"ACKSTREAM_STORAGE_HOSTS"`
	Keyspace string   `json:"keyspace" mapstructure:"ACKSTREAM_STORAGE_KEYSPACE"`
	Table    string   `json:"table" mapstructure:"ACKSTREAM_STORAGE_TABLE"`
}

func NewStorage(cfg *Configs) Storage {
	return &KVStorage{Configs: cfg}
}

type ctxkey string

const CTXKEY_CLIENT ctxkey = "ackstream.storage.client"

func FromContext(ctx context.Context) (Storage, error) {
	storage, ok := ctx.Value(CTXKEY_CLIENT).(Storage)
	if !ok {
		return nil, errors.New("no storage was configured")
	}

	return storage, nil
}
