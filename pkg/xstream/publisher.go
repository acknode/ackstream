package xstream

import (
	"context"
	"fmt"
	"strings"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func NewPub(ctx context.Context) (Pub, error) {
	logger := zlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("fn", "xstream.publisher")

	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return nil, ErrCfgNotSet
	}
	jsc, ok := StreamFromContext(ctx)
	if !ok {
		return nil, ErrStreamNotInit
	}

	return UsePub(cfg, jsc, logger), nil
}

func UsePub(cfg *Configs, streamctx nats.JetStreamContext, logger *zap.SugaredLogger) Pub {
	return func(e *entities.Event) (string, error) {
		msg := nats.NewMsg(NewSubject(cfg, e))
		msg.Data = []byte(e.Data)

		// with metadata
		msg.Header.Set("Nats-Msg-Id", e.Id)
		msg.Header.Set("AckStream-Event-Id", e.Id)
		msg.Header.Set("AckStream-Event-Bucket", e.Bucket)
		msg.Header.Set("AckStream-Event-Workspace", e.Workspace)
		msg.Header.Set("AckStream-Event-App", e.App)
		msg.Header.Set("AckStream-Event-Type", e.Type)
		msg.Header.Set("AckStream-Event-Creation-Time", fmt.Sprint(e.CreationTime))

		ack, err := streamctx.PublishMsg(msg)
		if err != nil {
			logger.Error(err.Error(), "key", e.Key())
			return "", err
		}

		keys := []string{
			ack.Domain, ack.Stream, fmt.Sprint(ack.Sequence), e.Id,
		}
		logger.Debugw("published", "stream_name", ack.Stream, "sequence", ack.Sequence, "key", e.Key())
		return strings.Join(keys, "/"), nil
	}
}
