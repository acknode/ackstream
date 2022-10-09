package xstorage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/zlogger"
)

type Put func(e *entities.Event) error

func UsePut(ctx context.Context, cfg *Configs) (Put, error) {
	logger := zlogger.FromContext(ctx).With("pkg", "storage", "fn", "storage.put")
	session, ok := ConnFromContext(ctx)
	if !ok {
		return nil, ErrConnNotInit
	}

	return func(e *entities.Event) error {
		// @TODO: validate event

		ql := fmt.Sprintf("INSERT INTO %s (bucket, workspace, app, type, id, data, creation_time) VALUES (?, ?, ?, ?, ?, ?, ?)", cfg.Table)

		// because we will set entities.Data type to interface{}, so we need to encode it as string when we insert to database
		data, err := json.Marshal(e.Data)
		if err != nil {
			return err
		}
		query := session.Query(ql, e.Bucket, e.Workspace, e.App, e.Type, e.Id, string(data), e.CreationTime)
		logger.Debugw("upsert entities", "ql", ql, "key", e.Key(), "data_length", len(data))

		return query.Exec()
	}, nil
}
