package xstream

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/logger"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func NewSub(ctx context.Context, cfg *Configs) Sub {
	l := logger.FromContext(ctx).
		With("pkg", "stream").
		With("fn", "stream.subscriber")

	// If stream was not initialized yet, we should init it
	conn, hasConn := ctx.Value(CTXKEY_CONN).(*nats.Conn)
	stream, hasStream := ctx.Value(CTXKEY_JS).(nats.JetStreamContext)
	if !hasConn || !hasStream {
		l.Debugw("no stream was provided, initialize a new one")
		stream, conn = New(ctx, cfg)
	}

	return func(topic, queue string, fn SubscribeFn) (func() error, error) {
		subject := NewSubject(cfg, topic, nil)
		l.Debugw("subscribed", "subject", subject, "queue", queue)

		sub, err := stream.QueueSubscribe(subject, queue, UseSub(fn, l))

		// return callback to cleanup resources
		return func() error {
			if err := sub.Unsubscribe(); err != nil {
				return err
			}

			conn.Close()
			return nil
		}, err
	}
}

func UseSub(fn SubscribeFn, l *zap.SugaredLogger) nats.MsgHandler {
	return func(msg *nats.Msg) {
		event := event.Event{
			Id:        msg.Header.Get("AckStream-Event-Id"),
			Bucket:    msg.Header.Get("AckStream-Event-Bucket"),
			Workspace: msg.Header.Get("AckStream-Event-Workspace"),
			App:       msg.Header.Get("AckStream-Event-App"),
			Type:      msg.Header.Get("AckStream-Event-Type"),
			Data:      msg.Data,
		}
		ll := l.With("key", event.Key())

		ct, err := strconv.ParseInt(msg.Header.Get("AckStream-Event-Creation-Time"), 10, 64)
		if err != nil {
			ll.Errorw(err.Error())
		}
		event.CreationTime = ct

		if err := fn(&event); err != nil {
			retry, _ := strconv.Atoi(msg.Header.Get("AckStream-Meta-Retry"))
			ll.Errorw(err.Error(), "retry", retry)

			msg.Header.Set("AckStream-Meta-Retry", fmt.Sprint(retry+1))
			// subcribers must handle error by themself
			// if they throw an error, message will be delivered again
			msg.NakWithDelay(time.Duration(math.Pow(2, float64(retry+1))))
			return
		}

		msg.Ack()
	}
}
