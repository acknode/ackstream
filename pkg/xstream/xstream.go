package xstream

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/gosimple/slug"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
)

type ctxkey string

const CTXKEY_CONN ctxkey = "ackstream.xstream.conn"
const CTXKEY_STREAM ctxkey = "ackstream.xstream.stream"

var MAX_MSG_SIZE int32 = 1024
var MAX_MSG int64 = 8192      // 8 * 1024
var MAX_BYTES int64 = 8388608 // 8 * 1024 * 1024
var MAX_AGE time.Duration = 3 * time.Hour

type Configs struct {
	Uri    string `json:"uri" mapstructure:"ACKSTREAM_STREAM_URI"`
	Region string `json:"region" mapstructure:"ACKSTREAM_STREAM_REGION"`
	Name   string `json:"name" mapstructure:"ACKSTREAM_STREAM_NAME"`
}

type SubscribeFn func(e *entities.Event) error

type Sub func(sample *entities.Event, queue string, fn SubscribeFn) (func() error, error)

type Pub func(e *entities.Event) (string, error)

func NewSubject(cfg *Configs, topic string, sample *entities.Event) string {
	segments := []string{cfg.Region, cfg.Name, topic}
	if sample == nil {
		return strings.Join(append(segments, ">"), ".")
	}

	if sample.Workspace != "" {
		segments = append(segments, sample.Workspace)
	} else {
		segments = append(segments, "*")
	}

	if sample.App != "" {
		segments = append(segments, sample.App)
	} else {
		segments = append(segments, "*")
	}

	if sample.Type != "" {
		segments = append(segments, sample.Type)
	} else {
		segments = append(segments, "*")
	}

	return strings.Join(segments, ".")
}

func New(ctx context.Context, cfg *Configs, topic string) (nats.JetStreamContext, *nats.Conn) {
	subjects := []string{strings.Join([]string{cfg.Region, cfg.Name, topic, ">"}, ".")}
	logger := zlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("stream_uri", cfg.Uri).
		With("stream_name", cfg.Name).
		With("stream_subjects", subjects)

	opts := []nats.Option{
		nats.ReconnectWait(3 * time.Second),
		nats.Timeout(3 * time.Second),
		nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
			// disconnected error could be nil, for instance when user explicitly closes the connection.
			if err != nil {
				logger.Errorw(err.Error())
			}
		}),
		nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
			logger.Errorw(err.Error(), "subject", s.Subject, "queue", s.Queue)
		}),
	}

	conn, err := nats.Connect(cfg.Uri, opts...)
	if err != nil {
		logger.Debugw(err.Error())
		panic(err)
	}

	jsc, err := conn.JetStream()
	if err != nil {
		logger.Debugw(err.Error())
		panic(err)
	}

	name := strings.ReplaceAll(slug.Make(cfg.Name), "-", "_")
	stream, err := jsc.StreamInfo(name)
	// if stream is exist, update the subject list
	if err == nil {
		stream.Config.Subjects = lo.Uniq(append(stream.Config.Subjects, subjects...))
		if stream, err = jsc.UpdateStream(&stream.Config); err != nil {
			logger.Debugw(err.Error())
			panic(err)
		}

		logger.Debug("updated existing stream")
	}

	// if there is no stream was created, create a new one
	if err != nil && errors.Is(err, nats.ErrStreamNotFound) {
		jscfg := nats.StreamConfig{
			Name:    name,
			Storage: nats.MemoryStorage,
			// replicas > 1 not supported in non-clustered mode
			// Replicas:  3,
			MaxMsgs:  MAX_MSG,
			MaxBytes: MAX_BYTES,
			MaxAge:   MAX_AGE,

			Subjects: subjects,
		}
		if stream, err = jsc.AddStream(&jscfg); err != nil {
			panic(err)
		}

		logger.Debug("created existing stream")
	}

	if stream == nil {
		logger.Debugw(err.Error())
		panic(err)
	}

	logger.Info("initialized stream successfully")
	return jsc, conn
}
