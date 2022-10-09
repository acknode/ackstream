package xstorage

import (
	"context"
	"fmt"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/zlogger"
)

type Scan func(sample *entities.Event, size int, page []byte) ([]entities.Event, []byte, error)

func UseScan(ctx context.Context, cfg *Configs) (Scan, error) {
	logger := zlogger.FromContext(ctx).With("pkg", "storage", "fn", "storage.scan")
	session, ok := ConnFromContext(ctx)
	if !ok {
		return nil, ErrConnNotInit
	}

	return func(sample *entities.Event, size int, page []byte) ([]entities.Event, []byte, error) {
		// @TODO: validate sample

		ql := fmt.Sprintf("SELECT id, data, creation_time FROM %s WHERE bucket = ? AND workspace = ? AND app = ? AND type = ? ORDER BY id DESC", cfg.Table)
		query := session.Query(ql, sample.Bucket, sample.Workspace, sample.App, sample.Type).PageSize(size)
		logger.Debugw("scan entitiess", "ql", ql, "size", size, "page", size)

		entitiess := []entities.Event{}
		iter := query.WithContext(ctx).PageState(page).Iter()
		scanner := iter.Scanner()

		for scanner.Next() {
			e := entities.Event{
				Bucket:    sample.Bucket,
				Workspace: sample.Workspace,
				App:       sample.App,
				Type:      sample.Type,
			}

			if err := scanner.Scan(&e.Id, &e.Data, &e.CreationTime); err != nil {
				iter.Close()
				return []entities.Event{}, nil, err
			}

			entitiess = append(entitiess, e)
		}

		// scanner.Err() closes the iterator, so scanner nor iter should be used afterwards.
		return entitiess, iter.PageState(), scanner.Err()
	}, nil
}
