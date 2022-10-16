package xstream

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"strings"
)

func NewPub(ctx context.Context) (Pub, error) {
	logger := xlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("fn", "xstream.pub")
	ctx = xlogger.WithContext(ctx, logger)

	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	jsc, err := NewJetStream(ctx)
	if err != nil {
		return nil, err
	}

	return func(event *entities.Event) (*string, error) {
		flogger := logger.With("event_key", event.Key())

		msg, err := NewMsg(cfg, event)
		if err != nil {
			flogger.Error(err.Error())
			return nil, err
		}

		ack, err := jsc.PublishMsg(msg)
		if err != nil {
			flogger.Error(err.Error())
			return nil, err
		}

		pubkey := strings.Join([]string{
			ack.Domain,
			ack.Stream,
			fmt.Sprint(ack.Sequence),
			event.Id,
		}, "/")
		logger.Debugw("published", "pubkey", pubkey)
		return &pubkey, nil
	}, nil
}
