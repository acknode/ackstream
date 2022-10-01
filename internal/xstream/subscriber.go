package xstream

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

func NewSub(ctx context.Context, cfg *Configs) Sub {
	logger := zlogger.FromContext(ctx).
		With("pkg", "stream").
		With("fn", "stream.subscriber")

	stream, _ := FromContext(ctx)

	return func(topic, queue string, fn SubscribeFn) (func() error, error) {
		subject := NewSubject(cfg, topic, nil)
		logger.Debugw("subscribed", "subject", subject, "queue", queue)

		sub, err := stream.QueueSubscribe(subject, queue, UseSub(fn, logger), nats.DeliverLast())

		// return callback to cleanup resources
		return func() error { return sub.Drain() }, err
	}
}

func UseSub(fn SubscribeFn, logger *zap.SugaredLogger) nats.MsgHandler {
	return func(msg *nats.Msg) {
		event := event.Event{
			Id:        msg.Header.Get("AckStream-Event-Id"),
			Bucket:    msg.Header.Get("AckStream-Event-Bucket"),
			Workspace: msg.Header.Get("AckStream-Event-Workspace"),
			App:       msg.Header.Get("AckStream-Event-App"),
			Type:      msg.Header.Get("AckStream-Event-Type"),
		}
		ll := logger.With("key", event.Key())
		if err := msgpack.Unmarshal(msg.Data, &event.Data); err != nil {
			ll.Error(err.Error())
			// if we could not decode the msg data, make sure we mark it as acknowledged
			return
		}

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
