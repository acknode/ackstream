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

func NewSub(ctx context.Context) (Sub, error) {
	logger := zlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("fn", "xstream.subscriber")

	ctx, err := Connect(ctx)
	if err != nil {
		return nil, err
	}
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return nil, ErrCfgNotSet
	}
	jsc, ok := StreamFromContext(ctx)
	if !ok {
		return nil, ErrStreamNotInit
	}

	return func(sample *entities.Event, queue string, fn SubscribeFn) (context.Context, error) {
		subject := NewSubject(cfg, sample)
		logger.Debugw("subscribed", "subject", subject, "queue", queue)

		sub, err := jsc.QueueSubscribe(subject, queue, UseSub(fn, logger), nats.DeliverLast())

		return SubcriptionWithContext(ctx, sub), err
	}, nil
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

func SubcriptionWithContext(ctx context.Context, sub *nats.Subscription) context.Context {
	return context.WithValue(ctx, CTXKEY_SUB, sub)
}

func SubcriptionFromContext(ctx context.Context) (*nats.Subscription, bool) {
	sub, ok := ctx.Value(CTXKEY_SUB).(*nats.Subscription)
	return sub, ok
}
