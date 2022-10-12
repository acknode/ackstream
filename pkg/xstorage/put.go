package xstorage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/gocql/gocql"
)

type Put func(e *entities.Event) error

func UsePut(ctx context.Context, cfg *Configs, session *gocql.Session) (Put, error) {
	logger := zlogger.FromContext(ctx).With("pkg", "storage", "fn", "storage.put")

	return func(e *entities.Event) error {
		// @TODO: validate event
		ql := fmt.Sprintf("INSERT INTO %s (bucket, workspace, app, type, id, data, timestamps) VALUES (?, ?, ?, ?, ?, ?, ?)", cfg.Table)

		// because we will set entities.Data type to interface{}, so we need to encode it as string when we insert to database
		data, err := json.Marshal(e.Data)
		if err != nil {
			return err
		}
		query := session.Query(ql, e.Bucket, e.Workspace, e.App, e.Type, e.Id, string(data), e.Timestamps)
		logger.Debugw("upsert entities", "ql", ql, "key", e.Key(), "data_length", len(data))

		return query.Exec()
	}, nil
}
