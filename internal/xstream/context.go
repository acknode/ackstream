package xstream

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
)

func WithContext(ctx context.Context, stream nats.JetStreamContext, conn *nats.Conn) context.Context {
	ctx = context.WithValue(ctx, CTXKEY_STREAM, stream)
	ctx = context.WithValue(ctx, CTXKEY_CONN, conn)
	return ctx
}

func FromContext(ctx context.Context) (stream nats.JetStreamContext, conn *nats.Conn) {
	stream, ok := ctx.Value(CTXKEY_STREAM).(nats.JetStreamContext)
	if !ok {
		panic(errors.New("no stream was configured"))
	}

	conn, ok = ctx.Value(CTXKEY_CONN).(*nats.Conn)
	if !ok {
		panic(errors.New("no stream connection was configured"))
	}

	return stream, conn
}
