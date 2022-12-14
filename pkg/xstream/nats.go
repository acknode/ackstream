package xstream

import (
	"context"
	"errors"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/gosimple/slug"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
	"os"
	"strings"
	"time"
)

const CTXKEY_CONN ctxkey = "ackstream.xstream.connection"

func ConnWithContext(ctx context.Context, conn *nats.Conn) context.Context {
	return context.WithValue(ctx, CTXKEY_CONN, conn)
}

func ConnFromContext(ctx context.Context) (*nats.Conn, error) {
	conn, ok := ctx.Value(CTXKEY_CONN).(*nats.Conn)
	if !ok {
		return nil, ErrConnNotFound
	}
	return conn, nil
}

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

	var dots []string
	count := len(segments)
	for i := count - 1; i >= 0; i-- {
		if segments[i] == "*" && segments[i-1] == "*" {
			continue
		}
		dots = append(dots, segments[i])
	}
	newdots := lo.Reverse[string](dots)
	if newdots[len(dots)-1] == "*" {
		newdots[len(dots)-1] = ">"
	}
	return strings.Join(newdots, ".")
}

func NewConnection(ctx context.Context) (*nats.Conn, error) {
	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}
	logger := xlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("xstream.uri", cfg.Uri).
		With("xstream.region", cfg.Region).
		With("xstream.name", cfg.Name).
		With("xstream.topic", cfg.Topic)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	opts := []nats.Option{
		nats.Name(hostname),
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
	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	subjects := []string{NewSubject(cfg, nil)}
	logger := xlogger.FromContext(ctx).
		With("pkg", "xstream").
		With("xstream.uri", cfg.Uri).
		With("xstream.region", cfg.Region).
		With("xstream.name", cfg.Name).
		With("xstream.topic", cfg.Topic).
		With("xstream.subjects", subjects)

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
			Name:     name,
			Replicas: 3,
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
