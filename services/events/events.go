package events

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events/protocol"
	"google.golang.org/grpc"
)

func New(ctx context.Context) (*grpc.Server, error) {
	pub, err := app.NewPub(ctx)
	if err != nil {
		return nil, err
	}

	logger := xlogger.FromContext(ctx)
	cfg := configs.FromContext(ctx)

	server := grpc.NewServer()
	protocol.RegisterEventsServer(server, &Server{logger: logger, cfg: cfg, pub: pub})
	return server, nil
}
