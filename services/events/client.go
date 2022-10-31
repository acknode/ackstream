package events

import (
	"context"
	"github.com/acknode/ackstream/services/events/protos"
	"google.golang.org/grpc"
)

func NewClient(ctx context.Context, conn *grpc.ClientConn) (protos.EventsClient, error) {
	return protos.NewEventsClient(conn), nil
}
