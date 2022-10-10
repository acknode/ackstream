package xstorage

import (
	"context"

	"github.com/gocql/gocql"
)

func New(ctx context.Context, cfg *Configs) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace

	return cluster.CreateSession()
}
