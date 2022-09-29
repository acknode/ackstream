package xstream

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/logger"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
)

type ctxkey string

const CTXKEY_CONN ctxkey = "ackstream.stream.conn"
const CTXKEY_JS ctxkey = "ackstream.stream.js"

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
		return strings.Join([]string{cfg.Region, cfg.Name, topic, ">"}, ".")
	}
	return strings.Join([]string{cfg.Region, cfg.Name, topic, e.Workspace, e.App, e.Type}, ".")
}

func New(ctx context.Context, cfg *Configs) (nats.JetStreamContext, *nats.Conn) {
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

	subject := fmt.Sprintf("%s.%s.>", cfg.Region, cfg.Name)
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

	return jsc, conn
}

func WithContext(ctx context.Context, conn *nats.Conn, js nats.JetStreamContext) context.Context {
	ctx = context.WithValue(ctx, CTXKEY_CONN, conn)
	ctx = context.WithValue(ctx, CTXKEY_JS, js)
	return ctx
}
