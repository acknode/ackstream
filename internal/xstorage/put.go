package xstorage

import (
	"context"
	"fmt"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/logger"
)

func UsePut(ctx context.Context, cfg *Configs) func(e *event.Event) error {
	l := logger.FromContext(ctx).With("pkg", "storage", "fn", "storage.put")
	session := FromContext(ctx)

	return func(e *event.Event) error {
		ql := fmt.Sprintf("INSERT INTO %s (bucket, workspace, app, type, id, data, creation_time) VALUES (?, ?, ?, ?, ?, ?, ?)", cfg.Table)
		query := session.Query(ql, e.Bucket, e.Workspace, e.App, e.Type, e.Id, e.Data, e.CreationTime)
		l.Debugw("upsert event", "ql", ql, "key", e.Key())

		return query.Exec()
	}
}
