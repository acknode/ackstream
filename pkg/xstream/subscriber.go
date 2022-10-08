package xstream

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func NewSub(ctx context.Context, cfg *Configs, topic string) Sub {
	logger := zlogger.FromContext(ctx).
		With("pkg", "stream").
		With("fn", "stream.subscriber")

	streamctx, _ := New(ctx, cfg, topic)

	return func(sample *entities.Event, queue string, fn SubscribeFn) (func() error, error) {
		subject := NewSubject(cfg, topic, sample)
		logger.Debugw("subscribed", "subject", subject, "queue", queue)

		sub, err := streamctx.QueueSubscribe(subject, queue, UseSub(fn, logger), nats.DeliverLast())

		// return callback to cleanup resources
		return func() error {
			if sub != nil {
				return sub.Drain()
			}
			return nil
		}, err
	}
}

func UseSub(fn SubscribeFn, logger *zap.SugaredLogger) nats.MsgHandler {
	return func(msg *nats.Msg) {
		entities := entities.Event{
			Id:        msg.Header.Get("AckStream-Event-Id"),
			Bucket:    msg.Header.Get("AckStream-Event-Bucket"),
			Workspace: msg.Header.Get("AckStream-Event-Workspace"),
			App:       msg.Header.Get("AckStream-Event-App"),
			Type:      msg.Header.Get("AckStream-Event-Type"),
			Data:      string(msg.Data),
		}
		ll := logger.With("key", entities.Key())
		ct, err := strconv.ParseInt(msg.Header.Get("AckStream-Event-Creation-Time"), 10, 64)
		if err != nil {
			ll.Errorw(err.Error())
		}
		entities.CreationTime = ct

		if err := fn(&entities); err != nil {
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
