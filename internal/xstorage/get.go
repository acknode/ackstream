package xstorage

import (
	"context"
	"fmt"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/pkg/zlogger"
)

func UseGet(ctx context.Context, cfg *Configs) func(bucket, workspace, app, etype string, id string) (*event.Event, error) {
	logger := zlogger.FromContext(ctx).With("pkg", "storage", "fn", "storage.get")
	session := FromContext(ctx)

	return func(bucket, workspace, app, etype string, id string) (*event.Event, error) {
		ql := fmt.Sprintf("SELECT data, creation_time FROM %s WHERE bucket = ? AND workspace = ? AND app = ? AND type = ? AND id = ?", cfg.Table)
		query := session.Query(ql, bucket, workspace, app, etype, id)
		logger.Debugw("scan events", "ql", ql, "id", id)

		e := event.Event{
			Bucket:    bucket,
			Workspace: workspace,
			App:       app,
			Type:      etype,
			Id:        id,
		}
		err := query.Scan(&e.Data, &e.CreationTime)
		logger.Debugw("get event", "ql", ql, "key", e.Key(), "found", err == nil)

		return &e, err
	}
}
