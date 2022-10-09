package xstorage

import (
	"context"

	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/gocql/gocql"
)

func NewConnection(ctx context.Context) (*gocql.Session, error) {
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return nil, ErrCfgNotSet
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace

	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}

	return session, nil
}

func Connect(ctx context.Context) (context.Context, error) {
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return ctx, ErrCfgNotSet
	}

	logger := zlogger.FromContext(ctx).
		With("pkg", "xstorage").
		With("xstorages.uri", cfg.Hosts).
		With("xstorages.keyspace", cfg.Keyspace).
		With("xstorages.table", cfg.Table).
		With("xstorages.bucket_template", cfg.BucketTemplate)

	xstoragectx := zlogger.WithContext(ctx, logger)
	conn, err := NewConnection(xstoragectx)
	if err != nil {
		logger.Debugw(err.Error())
		return ctx, err
	}
	ctx = ConnWithContext(ctx, conn)
	logger.Info("initialized connection successfully")

	return ctx, nil
}

func Disconnect(ctx context.Context) error {
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return ErrCfgNotSet
	}

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
