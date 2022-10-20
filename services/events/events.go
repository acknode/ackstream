package events

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	eventcfg "github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strings"
)

func New(ctx context.Context) (*http.Server, error) {
	pub, err := app.NewPub(ctx)
	if err != nil {
		return nil, err
	}

	logger := xlogger.FromContext(ctx)

	grpcServer := grpc.NewServer()
	proto.RegisterEventsServer(grpcServer, &Server{
		logger: logger,
		cfg:    configs.FromContext(ctx),
		pub:    pub,
	})

	cfg := eventcfg.FromContext(ctx)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = proto.RegisterEventsHandlerFromEndpoint(ctx, mux, cfg.ListenAddress, opts)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr: cfg.ListenAddress,
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				mux.ServeHTTP(w, r)
			}
		}), &http2.Server{}),
	}

	return srv, nil
}
