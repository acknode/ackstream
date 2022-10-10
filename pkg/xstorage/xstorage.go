package xstorage

import (
	"context"

	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/gocql/gocql"
)

func NewConnection(ctx context.Context, cfg *Configs) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace

	return cluster.CreateSession()
}

func Connect(ctx context.Context, cfg *Configs) (context.Context, error) {
	logger := zlogger.FromContext(ctx).
		With("pkg", "xstorage").
		With("xstorages.uri", cfg.Hosts).
		With("xstorages.keyspace", cfg.Keyspace).
		With("xstorages.table", cfg.Table).
		With("xstorages.bucket_template", cfg.BucketTemplate)

	conn, err := NewConnection(ctx, cfg)
	if err != nil {
		logger.Debugw(err.Error())
		return ctx, err
	}
	ctx = ConnWithContext(ctx, conn)
	logger.Info("initialized connection successfully")

	return ctx, nil
}

func Disconnect(ctx context.Context, cfg *Configs) error {
	logger := zlogger.FromContext(ctx).
		With("pkg", "xstorage").
		With("xstorages.uri", cfg.Hosts).
		With("xstorages.keyspace", cfg.Keyspace).
		With("xstorages.table", cfg.Table).
		With("xstorages.bucket_template", cfg.BucketTemplate)

	if conn, ok := ConnFromContext(ctx); ok {
		conn.Close()
		logger.Info("close connection successfully")
	}

	return nil
}
