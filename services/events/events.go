package events

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/services/events/protocol"
	"google.golang.org/grpc"
)

func New(ctx context.Context) (*grpc.Server, error) {
	pub, err := app.NewPub(ctx)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer()
	protocol.RegisterEventsServer(server, &Server{ctx: ctx, pub: pub})
	return server, nil
}
