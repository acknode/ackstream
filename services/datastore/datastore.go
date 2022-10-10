package datastore

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/acknode/ackstream/pkg/xstream"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/gocql/gocql"
)

type ctxkey string

const CTXKEY_QUEUE_NAME ctxkey = "ackstream.services.datastore.queue_name"

var ErrNoQueue = errors.New("stream queue name could not be empty")

func New(ctx context.Context, cfg *configs.Configs) error {
	queue, ok := ctx.Value(CTXKEY_QUEUE_NAME).(string)
	if !ok {
		return ErrNoQueue
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := zlogger.FromContext(ctx).With("service", "datastore")
	ctx = zlogger.WithContext(ctx, logger)

	conn, err := xstream.NewConnection(ctx, cfg.XStream)
	if err != nil {
		return err
	}
	jsc, err := xstream.NewJetStream(ctx, cfg.XStream, conn)
	if err != nil {
		return err
	}
	session, err := xstorage.New(ctx, cfg.XStorage)
	if err != nil {
		return err
	}

	go func() {
		sub := xstream.NewSub(ctx, cfg.XStream, jsc)
		handler, err := UseHandler(ctx, cfg, session)
		if err != nil {
			logger.Fatal(err.Error())
		}

		// because we don't provide a sample of event
		// so we will listen to all event changes
		if err := sub(nil, queue, handler); err != nil {
			logger.Fatal(err.Error())
		}

		logger.Debug("subscribing")
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()
	stop()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")
	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()

	go func() {
		time.Sleep(5 * time.Second)
		<-ctx.Done()
	}()

	if err := conn.Drain(); err != nil {
		return err
	}

	session.Close()

	return nil
}

func UseHandler(ctx context.Context, cfg *configs.Configs, session *gocql.Session) (xstream.SubscribeFn, error) {
	logger := zlogger.FromContext(ctx)
	put, err := xstorage.UsePut(ctx, cfg.XStorage, session)
	if err != nil {
		return nil, err
	}

	return func(e *entities.Event) error {
		logger.Debugw("got entities", "key", e.Key())
		return put(e)
	}, nil
}
