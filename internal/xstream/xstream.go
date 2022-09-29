package xstream

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/logger"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type ctxkey string

const CTXKEY_CLIENT ctxkey = "ackstream.stream.client"

var MAX_MSG_SIZE int32 = 1024
var MAX_MSG int64 = 8192      // 8 * 1024
var MAX_BYTES int64 = 8388608 // 8 * 1024 * 1024
var MAX_AGE time.Duration = 3 * time.Hour

type Configs struct {
	Uri    string `json:"uri" mapstructure:"ACKSTREAM_STREAM_URI"`
	Region string `json:"region" mapstructure:"ACKSTREAM_STREAM_REGION"`
	Name   string `json:"name" mapstructure:"ACKSTREAM_STREAM_NAME"`
}

type SubscribeFn func(e *event.Event) error

type Sub func(topic, queue string, fn SubscribeFn) (func() error, error)

type Pub func(topic string, e *event.Event) (string, error)

func NewSubject(cfg *Configs, topic string, e *event.Event) string {
	// if event is nill, that mean we want to subscribe all events from the partition that event is belong to
	if e == nil {
		return strings.Join([]string{cfg.Name, topic, ">"}, ".")
	}
	return strings.Join([]string{cfg.Name, topic, e.Workspace, e.App, e.Type}, ".")
}

func New(ctx context.Context, cfg *Configs) nats.JetStreamContext {
	l := logger.FromContext(ctx).With("pkg", "stream")

	opts := []nats.Option{
		nats.ReconnectWait(3 * time.Second),
		nats.Timeout(3 * time.Second),
		nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
			// disconnected error could be nil, for instance when user explicitly closes the connection.
			if err != nil {
				l.Errorw(err.Error())
			}
		}),
		nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
			l.Errorw(err.Error(), "subject", s.Subject, "queue", s.Queue)
		}),
	}

	conn, err := nats.Connect(cfg.Uri, opts...)
	if err != nil {
		l.Debugw(err.Error(), "uri", cfg.Uri)
		panic(err)
	}

	jsc, err := conn.JetStream()
	if err != nil {
		l.Debugw(err.Error(), "uri", cfg.Uri, "stream_name", cfg.Name)
		panic(err)
	}

	subject := fmt.Sprintf("%s.>", cfg.Name)
	stream, err := jsc.StreamInfo(cfg.Name)

	// if stream is exist, update the subject list
	if err == nil {
		stream.Config.Subjects = lo.Uniq(append(stream.Config.Subjects, subject))
		if stream, err = jsc.UpdateStream(&stream.Config); err != nil {
			l.Debugw(err.Error(),
				"uri", cfg.Uri,
				"stream_name", cfg.Name,
				"subject", subject,
			)
			panic(err)
		}
	}

	// if there is no stream was created, create a new one
	if err != nil && errors.Is(err, nats.ErrStreamNotFound) {
		jscfg := nats.StreamConfig{
			Name:    cfg.Name,
			Storage: nats.MemoryStorage,
			// replicas > 1 not supported in non-clustered mode
			// Replicas:  3,
			MaxMsgs:  MAX_MSG,
			MaxBytes: MAX_BYTES,
			MaxAge:   MAX_AGE,

			Subjects: []string{subject},
		}
		if stream, err = jsc.AddStream(&jscfg); err != nil {
			panic(err)
		}
	}

	if stream == nil {
		l.Debugw(err.Error(),
			"uri", cfg.Uri,
			"stream_name", cfg.Name,
			"subject", subject,
		)
		panic(errors.New("could not initialize stream successfully"))
	}

	return jsc
}

func NewPub(ctx context.Context, cfg *Configs) Pub {
	l := logger.FromContext(ctx).
		With("pkg", "stream").
		With("fn", "stream.publisher")

	// If stream was not initialized yet, we should init it
	stream, ok := ctx.Value(CTXKEY_CLIENT).(nats.JetStreamContext)
	if !ok {
		l.Debugw("no stream was provided, initialize a new one")
		stream = New(ctx, cfg)
	}

	return func(topic string, e *event.Event) (string, error) {
		msg := nats.NewMsg(NewSubject(cfg, topic, e))
		msg.Data = e.Data

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

func NewSub(ctx context.Context, cfg *Configs) Sub {
	l := logger.FromContext(ctx).
		With("pkg", "stream").
		With("fn", "stream.subscriber")

	// If stream was not initialized yet, we should init it
	stream, ok := ctx.Value(CTXKEY_CLIENT).(nats.JetStreamContext)
	if !ok {
		l.Debugw("no stream was provided, initialize a new one")
		stream = New(ctx, cfg)
	}

	return func(topic, queue string, fn SubscribeFn) (func() error, error) {
		subject := NewSubject(cfg, topic, nil)

		sub, err := stream.QueueSubscribe(subject, queue, UseSub(fn, l))

		// return callback to cleanup resources
		return func() error { return sub.Drain() }, err
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
