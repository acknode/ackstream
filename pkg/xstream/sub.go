package xstream

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
	"math"
	"strconv"
	"time"
)

func NewSub(ctx context.Context) (Sub, error) {
	logger := xlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("fn", "xstream.sub")
	ctx = xlogger.WithContext(ctx, logger)

	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	jsc, err := NewJetStream(ctx)
	if err != nil {
		return nil, err
	}

	return func(sample *entities.Event, queue string, fn SubscribeFn) error {
		if queue == "" {
			return ErrSubNoQueue
		}

		subject := NewSubject(cfg, sample)

		opts := map[string]nats.SubOpt{"delivery": nats.DeliverNew()}
		if cfg.ConsumerPolicy == CONSUMER_POLICY_ALL {
			opts["delivery"] = nats.DeliverAll()
		}

		// by default the consumer that is created by QueueSubscribe will be there forever (set durable to TRUE)
		if _, err := jsc.QueueSubscribe(subject, queue, UseSub(ctx, fn), lo.Values(opts)...); err != nil {
			logger.Errorw(err.Error(), "subject", subject, "queue", queue)
			return err
		}

		logger.Debugw("subscribed", "subject", subject, "queue", queue)
		return nil
	}, nil
}

func UseSub(ctx context.Context, fn SubscribeFn) nats.MsgHandler {
	logger := xlogger.FromContext(ctx)

	return func(msg *nats.Msg) {
		event, err := NewEvent(msg)
		if err != nil {
			logger.Error(err)
			if err := msg.Ack(); err != nil {
				logger.Errorw("ack was failed", "error", err.Error())
			}
			return
		}
		flogger := logger.With("event_key", event.Key())
		flogger.Debug("got event")

		if err := fn(event); err != nil {
			retry, _ := strconv.ParseInt(msg.Header.Get("AckStream-Meta-Retry"), 10, 64)
			flogger.Errorw("could not handle event", "error", err.Error(), "retry_count", retry)

			msg.Header.Set("AckStream-Meta-Retry", fmt.Sprint(retry+1))
			// subscribers must handle error by themselves
			// if they throw an error, message will be delivered again
			if err := msg.NakWithDelay(time.Duration(math.Pow(2, float64(retry+1)))); err != nil {
				logger.Errorw("nak was failed", "error", err.Error())
			}
			return
		}

		if err := msg.Ack(); err != nil {
			logger.Errorw("ack was failed", "error", err.Error())
		}
		flogger.Debug("processed event")
	}
}
