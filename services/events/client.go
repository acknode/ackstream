package events

import (
	"context"
	"github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/protos"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(ctx context.Context) (*grpc.ClientConn, protos.EventsClient, error) {
	cfg := configs.FromContext(ctx)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(grpcRetry.StreamClientInterceptor(grpcRetry.WithMax(3))),
		grpc.WithUnaryInterceptor(grpcRetry.UnaryClientInterceptor(grpcRetry.WithMax(3))),
	}
	conn, err := grpc.Dial(cfg.GRPCListenAddress, opts...)
	if err != nil {
		return nil, nil, err
	}

	return conn, protos.NewEventsClient(conn), nil
}
