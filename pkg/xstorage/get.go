package xstorage

import (
	"context"
	"fmt"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/gocql/gocql"
)

type Get func(sample *entities.Event) (*entities.Event, error)

func UseGet(ctx context.Context, cfg *Configs, session *gocql.Session) (Get, error) {
	logger := zlogger.FromContext(ctx).With("pkg", "storage", "fn", "storage.get")

	return func(sample *entities.Event) (*entities.Event, error) {
		ql := fmt.Sprintf("SELECT data, creation_time FROM %s WHERE bucket = ? AND workspace = ? AND app = ? AND type = ? AND id = ?", cfg.Table)
		query := session.Query(ql, sample.Bucket, sample.Workspace, sample.App, sample.Type, sample.Id)
		logger.Debugw("scan entitiess", "ql", ql, "id", sample.Id)

		e := entities.Event{
			Bucket:    sample.Bucket,
			Workspace: sample.Workspace,
			App:       sample.App,
			Type:      sample.Type,
			Id:        sample.Id,
		}
		err := query.Scan(&e.Data, &e.CreationTime)
		logger.Debugw("get entities", "ql", ql, "key", e.Key(), "found", err == nil)

		return &e, err
	}, nil
}
