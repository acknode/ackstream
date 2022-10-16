package xstream

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/nats-io/nats.go"
	"math"
	"strconv"
	"time"
)

func NewSub(ctx context.Context) (Sub, error) {
	logger := xlogger.FromContext(ctx).
		With("pkg", "xstream")

	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	jsc, err := NewJetStream(ctx)
	if err != nil {
		return nil, err
	}

	return func(sample *entities.Event, queue string, fn SubscribeFn) error {
		subject := NewSubject(cfg, sample)
		_, err := jsc.QueueSubscribe(subject, queue, UseSub(ctx, fn), nats.DeliverLast())

		logger.Debugw("subscribed", "subject", subject, "queue", queue)
		return err
	}, nil

}

func UseSub(ctx context.Context, fn SubscribeFn) nats.MsgHandler {
	logger := xlogger.FromContext(ctx).
		With("fn", "xstream.subscriber")

	return func(msg *nats.Msg) {
		event, err := NewEvent(msg)
		if err != nil {
			logger.Error(err)
			if err := msg.Ack(); err != nil {
				logger.Errorw("ack was failed", "error", err.Error())
			}
			return
		}
		fnlogger := logger.With("event_key", event.Key())
		fnlogger.Debug("got event")

		if err := fn(event); err != nil {
			retry, _ := strconv.ParseInt(msg.Header.Get("AckStream-Meta-Retry"), 10, 64)
			fnlogger.Errorw("could not handle event", "error", err.Error(), "retry_count", retry)

			msg.Header.Set("AckStream-Meta-Retry", fmt.Sprint(retry+1))
			// subcribers must handle error by themselves
			// if they throw an error, message will be delivered again
			if err := msg.NakWithDelay(time.Duration(math.Pow(2, float64(retry+1)))); err != nil {
				logger.Errorw("nak was failed", "error", err.Error())
			}
			return
		}

		if err := msg.Ack(); err != nil {
			logger.Errorw("ack was failed", "error", err.Error())
		}
		fnlogger.Debug("processed event")
	}
}
