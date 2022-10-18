package xstorage

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
)

type Put func(event *entities.Event) error

func NewPut(ctx context.Context) (Put, error) {
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

	return func(event *entities.Event) error {
		if !event.Valid() {
			return ErrEventInvalid
		}

		ql := fmt.Sprintf("INSERT INTO %s (bucket, workspace, app, type, id, data, timestamps) VALUES (?, ?, ?, ?, ?, ?, ?)", cfg.Table)
		flogger := logger.With("event_key", event.Key(), "ql", ql)

		query := conn.Query(ql,
			event.Bucket, event.Workspace, event.App, event.Type, event.Id,
			event.Data, event.Timestamps,
		)
		flogger.Debugw("put event", "data_length", len(event.Data))

		return query.Exec()
	}, nil
}
