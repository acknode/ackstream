package datastore

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/acknode/ackstream/pkg/xstream"
	"github.com/acknode/ackstream/pkg/zlogger"
)

type ctxkey string

const CTXKEY_QUEUE_NAME ctxkey = "ackstream.services.datastore.queue_name"

var ErrNoQueue = errors.New("stream queue name could not be empty")

func New(ctx context.Context) error {
	queue, ok := ctx.Value(CTXKEY_QUEUE_NAME).(string)
	if !ok {
		return ErrNoQueue
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := zlogger.FromContext(ctx).With("service", "datastore")
	ctx = zlogger.WithContext(ctx, logger)

	var err error
	var cleanup func() error
	go func() {
		sub := app.UseSub(ctx)
		// because we don't provide a sample of event
		// so we will listen to all event changes
		cleanup, err = sub(nil, queue, UseHandler(ctx))
		if err != nil {
			logger.Fatal(err.Error())
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()
	stop()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")
	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cleanup(); err != nil {
		logger.Fatal(err.Error())
	}

	return nil
}

func UseHandler(ctx context.Context) xstream.SubscribeFn {
	cfg := configs.FromContext(ctx)
	logger := zlogger.FromContext(ctx)
	put := xstorage.UsePut(ctx, cfg.XStorage)

	return func(e *entities.Event) error {
		logger.Debugw("got entities", "key", e.Key())
		return put(e)
	}
}
