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

	ctx, err := xstream.Connect(ctx)
	if err != nil {
		return err
	}

	ctx, err = xstorage.Connect(ctx)
	if err != nil {
		return err
	}

	go func() {
		sub, err := xstream.NewSub(ctx)
		if err != nil {
			logger.Fatal(err.Error())
		}

		handler, err := UseHandler(ctx)
		if err != nil {
			logger.Fatal(err.Error())
		}

		// because we don't provide a sample of event
		// so we will listen to all event changes
		ctx, err = sub(nil, queue, handler)
		if err != nil {
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := xstream.Disconnect(ctx); err != nil {
		return err
	}

	if err := xstorage.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func UseHandler(ctx context.Context) (xstream.SubscribeFn, error) {
	cfg := configs.FromContext(ctx)
	logger := zlogger.FromContext(ctx)
	put, err := xstorage.UsePut(ctx, cfg.XStorage)
	if err != nil {
		return nil, err
	}

	return func(e *entities.Event) error {
		logger.Debugw("got entities", "key", e.Key())
		return put(e)
	}, nil
}
