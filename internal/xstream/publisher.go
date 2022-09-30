package xstream

import (
	"context"
	"fmt"
	"strings"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/zlogger"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

func NewPub(ctx context.Context, cfg *Configs) Pub {
	logger := zlogger.FromContext(ctx).
		With("pkg", "stream").
		With("fn", "stream.publisher")

	stream, _ := FromContext(ctx)
	return UsePub(cfg, stream, logger)
}

func UsePub(cfg *Configs, stream nats.JetStreamContext, l *zap.SugaredLogger) Pub {
	return func(topic string, e *event.Event) (string, error) {
		msg := nats.NewMsg(NewSubject(cfg, topic, e))

		data, err := msgpack.Marshal(e.Data)
		if err != nil {
			return "", err
		}
		msg.Data = data

		// with metadata
		msg.Header.Set("Nats-Msg-Id", e.Id)
		msg.Header.Set("AckStream-Event-Id", e.Id)
		msg.Header.Set("AckStream-Event-Bucket", e.Bucket)
		msg.Header.Set("AckStream-Event-Workspace", e.Workspace)
		msg.Header.Set("AckStream-Event-App", e.App)
		msg.Header.Set("AckStream-Event-Type", e.Type)
		msg.Header.Set("AckStream-Event-Creation-Time", fmt.Sprint(e.CreationTime))

		ack, err := stream.PublishMsg(msg)
		if err != nil {
			l.Error(err.Error(), "key", e.Key())
			return "", err
		}

		keys := []string{
			ack.Domain, ack.Stream, fmt.Sprint(ack.Sequence), e.Id,
		}
		l.Debugw("published", "stream_name", ack.Stream, "sequence", ack.Sequence, "key", e.Key())
		return strings.Join(keys, "/"), nil
	}
}
