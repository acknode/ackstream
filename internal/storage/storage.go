package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/gocql/gocql"
)

func New(cfg *Configs) *Storage {
	return &Storage{Configs: cfg}
}

// use Scylla - a alternative version of Cassandra for key-value storage
type Storage struct {
	Configs *Configs
	Cluster *gocql.ClusterConfig
	Session *gocql.Session
}

func (storage *Storage) Start() error {
	// support start multiple times
	if storage.Session != nil && !storage.Session.Closed() {
		return nil
	}

	storage.Cluster = gocql.NewCluster(storage.Configs.Hosts...)
	storage.Cluster.Keyspace = storage.Configs.Keyspace

	session, err := storage.Cluster.CreateSession()
	if err != nil {
		return err
	}

	storage.Session = session

	if ok := Ping(storage); !ok {
		return errors.New("could not connect to storage")
	}
	return nil
}

func (storage *Storage) Stop() error {
	if storage.Session != nil && !storage.Session.Closed() {
		storage.Session.Close()
	}

	storage.Cluster = nil
	storage.Session = nil
	return nil
}

func (storage *Storage) Put(ctx context.Context, e *event.Event) error {
	ql := fmt.Sprintf("INSERT INTO %s (bucket, workspace, app, type, id, data, creation_time) VALUES (?, ?, ?, ?, ?, ?, ?)", storage.Configs.Table)
	query := storage.Session.Query(ql, e.Bucket, e.Workspace, e.App, e.Type, e.Id, e.Data, e.CreationTime)

	newctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	return query.WithContext(newctx).Exec()
}

func (storage *Storage) Get(ctx context.Context, bucket, workspace, app, msgtype string, id string) (*event.Event, error) {
	ql := fmt.Sprintf("SELECT payload, creation_time FROM %s WHERE bucket = ? AND workspace = ? AND app = ? AND type = ? AND id = ?", storage.Configs.Table)
	query := storage.Session.Query(ql, bucket, workspace, app, msgtype, id)

	newctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	event := event.Event{
		Bucket:    bucket,
		Workspace: workspace,
		App:       app,
		Type:      msgtype,
		Id:        id,
	}
	err := query.WithContext(newctx).Scan(&event.Data, &event.CreationTime)
	return &event, err
}

func (storage *Storage) Scan(ctx context.Context, bucket, workspace, app, msgtype string, size int, page []byte) ([]event.Event, []byte, []error) {
	ql := fmt.Sprintf("SELECT id, payload, creation_time FROM %s WHERE bucket = ? AND workspace = ? AND app = ? AND type = ? ORDER BY id DESC", storage.Configs.Table)
	query := storage.Session.Query(ql, bucket, workspace, app, msgtype).PageSize(size)

	events := []event.Event{}
	errs := []error{}

	iter := query.WithContext(ctx).PageState(page).Iter()
	scanner := iter.Scanner()
	for scanner.Next() {
		event := event.Event{
			Bucket:    bucket,
			Workspace: workspace,
			App:       app,
			Type:      msgtype,
		}

		if err := scanner.Scan(&event.Id, &event.Data, &event.CreationTime); err != nil {
			errs = append(errs, err)
			continue
		}

		events = append(events, event)
	}

	// scanner.Err() closes the iterator, so scanner nor iter should be used afterwards.
	if err := scanner.Err(); err != nil {
		errs = append(errs, err)
	}

	return events, iter.PageState(), errs
}
