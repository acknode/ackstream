package xstorage

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/vmihailenco/msgpack/v5"
)

type Get func(sample *entities.Event) (*entities.Event, error)

func NewGet(ctx context.Context) (Get, error) {
	logger := xlogger.FromContext(ctx).
		With("pkg", "xstorage").
		With("fn", "put")

	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return func(sample *entities.Event) (*entities.Event, error) {
		event := entities.Event{
			Bucket:    sample.Bucket,
			Workspace: sample.Workspace,
			App:       sample.App,
			Type:      sample.Type,
			Id:        sample.Id,
		}
		if !event.Valid() {
			return nil, ErrEventQueryInvalid
		}

		ql := fmt.Sprintf("SELECT data, timestamps FROM %s WHERE bucket = ? AND workspace = ? AND app = ? AND type = ? AND id = ?", cfg.Table)
		flogger := logger.With("event_key", event.Key(), "ql", ql)

		query := conn.Query(ql, event.Bucket, event.Workspace, event.App, event.Type, event.Id)

		var data []byte
		err := query.Scan(&data, &event.Timestamps)
		if err != nil {
			return nil, err
		}
		flogger.Debugw("get entities", "ql", ql, "key", event.Key(), "found", err == nil)

		if err := msgpack.Unmarshal(data, &event.Data); err != nil {
			return nil, err
		}

		return &event, err
	}, nil
}
