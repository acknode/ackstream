package events

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	eventcfg "github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/proto"
	"github.com/acknode/ackstream/utils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

func New(ctx context.Context) (*http.Server, error) {
	pub, err := app.NewPub(ctx)
	if err != nil {
		return nil, err
	}

	logger := xlogger.FromContext(ctx)
	cfg := eventcfg.FromContext(ctx)

	grpcServer, err := utils.NewGRPCServer(cfg.CertsDir)
	if err != nil {
		return nil, err
	}

	proto.RegisterEventsServer(grpcServer, &Server{
		logger: logger,
		cfg:    configs.FromContext(ctx),
		pub:    pub,
	})

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = proto.RegisterEventsHandlerFromEndpoint(ctx, mux, cfg.ListenAddress, opts)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr: cfg.ListenAddress,
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isGRPC := r.Header.Get("Content-Type") == "application/grpc"
			if r.ProtoMajor == 2 && isGRPC {
				grpcServer.ServeHTTP(w, r)
			} else {
				mux.ServeHTTP(w, r)
			}
		}), &http2.Server{}),
	}

	return srv, nil
}
