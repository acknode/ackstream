package app

import (
	"context"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/acknode/ackstream/pkg/xstream"
)

func Connect(ctx context.Context) (context.Context, error) {
	stream, err := xstream.NewConnection(ctx)
	if err != nil {
		return ctx, err
	}
	ctx = xstream.ConnWithContext(ctx, stream)

	storage, err := xstorage.NewConnection(ctx)
	if err != nil {
		return ctx, err
	}
	ctx = xstorage.ConnWithContext(ctx, storage)

	return ctx, nil
}

func Disconnect(ctx context.Context) (context.Context, error) {
	logger := xlogger.FromContext(ctx)
	if stream, err := xstream.ConnFromContext(ctx); err == nil {
		if err := stream.Drain(); err != nil {
			logger.Errorw(err.Error(), "pkg", "xstream")
		}
	}

	if storage, err := xstorage.ConnFromContext(ctx); err == nil {
		storage.Close()
	}

	return ctx, nil
}
