package xstream

import (
	"context"

	"github.com/nats-io/nats.go"
)

type ctxkey string

const CTXKEY_CONN ctxkey = "xstream.conn"
const CTXKEY_STREAM ctxkey = "xstream.stream"
const CTXKEY_SUB ctxkey = "xstream.subscription"

func ConnWithContext(ctx context.Context, conn *nats.Conn) context.Context {
	return context.WithValue(ctx, CTXKEY_CONN, conn)
}
func ConnFromContext(ctx context.Context) (*nats.Conn, bool) {
	conn, ok := ctx.Value(CTXKEY_CONN).(*nats.Conn)
	return conn, ok
}

func StreamWithContext(ctx context.Context, jsc nats.JetStreamContext) context.Context {
	return context.WithValue(ctx, CTXKEY_STREAM, jsc)
}
func StreamFromContext(ctx context.Context) (nats.JetStreamContext, bool) {
	jsc, ok := ctx.Value(CTXKEY_STREAM).(nats.JetStreamContext)
	return jsc, ok
}
