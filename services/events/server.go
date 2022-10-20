package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	eventcfg "github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"net/http"
	"os"
)

func NewGRPCServer(ctx context.Context) (*grpc.Server, error) {
	pub, err := app.NewPub(ctx)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterEventsServer(grpcServer, &Server{
		logger: xlogger.FromContext(ctx),
		cfg:    configs.FromContext(ctx),
		pub:    pub,
	})
	reflection.Register(grpcServer)

	return grpcServer, nil
}

func NewHTTPServer(ctx context.Context) (*http.Server, error) {
	cfg := eventcfg.FromContext(ctx)

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := proto.RegisterEventsHandlerFromEndpoint(ctx, mux, cfg.GRPCListenAddress, opts)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr: cfg.HTTPListenAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mux.ServeHTTP(w, r)
		}),
	}

	return srv, nil
}

type Server struct {
	proto.EventsServer

	logger *zap.SugaredLogger
	cfg    *configs.Configs
	pub    app.Pub
}

func (s *Server) Pub(ctx context.Context, req *proto.PubReq) (*proto.PubRes, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		s.logger.Debugw("got metadata", "meta", meta)
	}

	event := &entities.Event{
		Workspace: req.Workspace,
		App:       req.App,
		Type:      req.Type,
		Data:      req.Data,
	}
	if err := event.WithBucket(s.cfg.BucketTemplate); err != nil {
		s.logger.Error(err)
		return nil, err
	}
	s.logger.Debugw("got events", "event_key", event.Key(), "data_length", len(req.Data))

	if err := event.WithId(); err != nil {
		s.logger.Error(err)
		return nil, err
	}

	if !event.Valid() {
		msg := fmt.Sprintf("services.events: %s is not valid event", event.Key())
		s.logger.Error(msg)
		return nil, errors.New(msg)
	}

	pubkey, err := s.pub(event)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	res := &proto.PubRes{
		Pubkey:     *pubkey,
		Bucket:     event.Bucket,
		Timestamps: event.Timestamps,
	}
	return res, nil
}

func (s *Server) Health(context.Context, *proto.HealthReq) (*proto.HealthRes, error) {
	host, _ := os.Hostname()
	res := &proto.HealthRes{
		Host:    host,
		Version: s.cfg.Version,
	}
	return res, nil
}
