package xstorage

import (
	"context"

	"github.com/gocql/gocql"
)

type ctxkey string

const CTXKEY_CONN ctxkey = "ackstream.storage.conn"

func ConnWithContext(ctx context.Context, session *gocql.Session) context.Context {
	return context.WithValue(ctx, CTXKEY_CONN, session)
}

func ConnFromContext(ctx context.Context) (*gocql.Session, bool) {
	session, ok := ctx.Value(CTXKEY_CONN).(*gocql.Session)
	return session, ok
}
