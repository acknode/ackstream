package events

import (
	"context"
	"github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(ctx context.Context) (*grpc.ClientConn, protos.EventsClient, error) {
	cfg := configs.FromContext(ctx)

	transportOpts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(cfg.GRPCListenAddress, transportOpts)
	if err != nil {
		return nil, nil, err
	}

	return conn, protos.NewEventsClient(conn), nil
}
