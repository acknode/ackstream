package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events/protos"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"os"
)

func NewServer(ctx context.Context) (*grpc.Server, error) {
	pub, err := app.NewPub(ctx)
	if err != nil {
		return nil, err
	}

	logger := xlogger.FromContext(ctx)
	cfg := configs.FromContext(ctx)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				logger.Debugw("handling request", "request.method", info.FullMethod)
				resp, err = handler(ctx, req)

				return
			},
		),
	)
	protos.RegisterEventsServer(server, &Server{
		logger: logger,
		cfg:    cfg,
		pub:    pub,
	})
	reflection.Register(server)

	return server, nil
}

type Server struct {
	protos.EventsServer

	logger *zap.SugaredLogger
	cfg    *configs.Configs
	pub    app.Pub
}

func (s *Server) Pub(ctx context.Context, req *protos.PubReq) (*protos.PubRes, error) {
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

	res := &protos.PubRes{
		Pubkey:     *pubkey,
		Bucket:     event.Bucket,
		Timestamps: event.Timestamps,
	}
	return res, nil
}

func (s *Server) Health(context.Context, *protos.HealthReq) (*protos.HealthRes, error) {
	host, _ := os.Hostname()
	res := &protos.HealthRes{
		Host:    host,
		Version: s.cfg.Version,
	}
	return res, nil
}
