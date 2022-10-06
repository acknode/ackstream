package xstorage

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
)

type ctxkey string

const CTXKEY_SESSION ctxkey = "ackstream.storage.session"

type Configs struct {
	Hosts          []string `json:"hosts" mapstructure:"ACKSTREAM_STORAGE_HOSTS"`
	Keyspace       string   `json:"keyspace" mapstructure:"ACKSTREAM_STORAGE_KEYSPACE"`
	Table          string   `json:"table" mapstructure:"ACKSTREAM_STORAGE_TABLE"`
	BucketTemplate string   `json:"bucket_template" mapstructure:"ACKSTREAM_STORAGE_BUCKET_TEMPLATE"`
}

func New(ctx context.Context, cfg *Configs) *gocql.Session {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace

	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}

	return session
}

func WithContext(ctx context.Context, session *gocql.Session) context.Context {
	return context.WithValue(ctx, CTXKEY_SESSION, session)
}

func FromContext(ctx context.Context) *gocql.Session {
	session, ok := ctx.Value(CTXKEY_SESSION).(*gocql.Session)
	if !ok {
		panic(errors.New("no storage session was configured"))
	}

	return session
}
