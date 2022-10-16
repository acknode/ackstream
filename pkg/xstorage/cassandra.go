package xstorage

import (
	"context"
	"github.com/gocql/gocql"
)

const CTXKEY_CONN ctxkey = "ackstream.xstorage.connection"

func ConnWithContext(ctx context.Context, conn *gocql.Session) context.Context {
	return context.WithValue(ctx, CTXKEY_CONN, conn)
}

func ConnFromContext(ctx context.Context) (*gocql.Session, error) {
	conn, ok := ctx.Value(CTXKEY_CONN).(*gocql.Session)
	if !ok {
		return nil, ErrConnNotFound
	}
	return conn, nil
}

func NewConnection(ctx context.Context) (*gocql.Session, error) {
	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace

	return cluster.CreateSession()
}
