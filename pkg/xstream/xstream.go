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

var MAX_MSG int64 = 8192      // 8 * 1024
var MAX_BYTES int64 = 8388608 // 8 * 1024 * 1024
var MAX_AGE time.Duration = 3 * time.Hour

type SubscribeFn func(e *entities.Event) error

type Sub func(sample *entities.Event, queue string, fn SubscribeFn) (context.Context, error)

type Pub func(e *entities.Event) (string, error)

func NewSubject(cfg *Configs, sample *entities.Event) string {
	segments := []string{cfg.Region, cfg.Name, cfg.Topic}
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

func NewConnection(ctx context.Context) (*nats.Conn, error) {
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return nil, ErrCfgNotSet
	}
	logger := zlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("xstream.uri", cfg.Uri).
		With("xstream.region", cfg.Region).
		With("xstream.name", cfg.Name).
		With("xstream.topic", cfg.Topic)

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

	return nats.Connect(cfg.Uri, opts...)
}

func NewJetStream(ctx context.Context) (nats.JetStreamContext, error) {
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return nil, ErrCfgNotSet
	}
	subjects := []string{NewSubject(cfg, nil)}
	logger := zlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("xstream.uri", cfg.Uri).
		With("xstream.region", cfg.Region).
		With("xstream.name", cfg.Name).
		With("xstream.topic", cfg.Topic).
		With("xstream.subjects", subjects)

	conn, ok := ConnFromContext(ctx)
	if !ok {
		return nil, ErrConnNotInit
	}

	jsc, err := conn.JetStream()
	if err != nil {
		return nil, err
	}

	name := strings.ReplaceAll(slug.Make(cfg.Name), "-", "_")
	stream, err := jsc.StreamInfo(name)
	// if stream is exist, update the subject list
	if err == nil {
		stream.Config.Subjects = lo.Uniq(append(stream.Config.Subjects, subjects...))
		if _, err = jsc.UpdateStream(&stream.Config); err != nil {
			return nil, err
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
			MaxMsgs:  cfg.MaxMsgs,
			MaxBytes: cfg.MaxBytes,
			MaxAge:   time.Duration(cfg.MaxAge) * time.Hour,

			Subjects: subjects,
		}
		if _, err = jsc.AddStream(&jscfg); err != nil {
			return nil, err
		}

		logger.Debug("created new stream")
	}

	return jsc, err
}

func Connect(ctx context.Context) (context.Context, error) {
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return ctx, ErrCfgNotSet
	}

	logger := zlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("xstream.uri", cfg.Uri).
		With("xstream.region", cfg.Region).
		With("xstream.name", cfg.Name).
		With("xstream.topic", cfg.Topic)

	conn, err := NewConnection(ctx)
	if err != nil {
		logger.Debugw(err.Error())
		return ctx, err
	}
	ctx = ConnWithContext(ctx, conn)
	logger.Info("initialized connection successfully")

	jsc, err := NewJetStream(ctx)
	if err != nil {
		logger.Debugw(err.Error())
		return ctx, err
	}

	ctx = StreamWithContext(ctx, jsc)
	logger.Info("initialized stream successfully")

	return ctx, nil
}

func Disconnect(ctx context.Context) error {
	cfg, ok := CfgFromContext(ctx)
	if !ok {
		return ErrCfgNotSet
	}

	logger := zlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("xstream.uri", cfg.Uri).
		With("xstream.region", cfg.Region).
		With("xstream.name", cfg.Name).
		With("xstream.topic", cfg.Topic)

	if conn, ok := ConnFromContext(ctx); ok {
		conn.Drain()
		logger.Info("drain connection successfully")
	}

	if subscription, ok := SubcriptionFromContext(ctx); ok {
		if err := subscription.Drain(); err != nil {
			return err
		}
		logger.Info("drain subscription successfully")
	}

	return nil
}
