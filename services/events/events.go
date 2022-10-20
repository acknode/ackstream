package events

import (
	"context"
	"google.golang.org/grpc"
	"net/http"
)

func NewServers(ctx context.Context) (*grpc.Server, *http.Server, error) {
	gRPCServer, err := NewGRPCServer(ctx)
	if err != nil {
		return nil, nil, err
	}

	httpServer, err := NewHTTPServer(ctx)
	if err != nil {
		return nil, nil, err
	}

	return gRPCServer, httpServer, nil
}
