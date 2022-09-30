package xstorage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/zlogger"
)

func UsePut(ctx context.Context, cfg *Configs) func(e *event.Event) error {
	logger := zlogger.FromContext(ctx).With("pkg", "storage", "fn", "storage.put")
	session := FromContext(ctx)

	return func(e *event.Event) error {
		ql := fmt.Sprintf("INSERT INTO %s (bucket, workspace, app, type, id, data, creation_time) VALUES (?, ?, ?, ?, ?, ?, ?)", cfg.Table)

		// because we will set event.Data type to interface{}, so we need to encode it as string when we insert to database
		data, err := json.Marshal(e.Data)
		if err != nil {
			return err
		}
		query := session.Query(ql, e.Bucket, e.Workspace, e.App, e.Type, e.Id, string(data), e.CreationTime)
		logger.Debugw("upsert event", "ql", ql, "key", e.Key(), "data_length", len(data))

		return query.Exec()
	}
}
