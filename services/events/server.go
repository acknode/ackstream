package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/services/events/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"os"
)

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

func (s *Server) Health(context.Context, *proto.PingReq) (*proto.PingRes, error) {
	host, _ := os.Hostname()
	res := &proto.PingRes{
		Host:    host,
		Version: s.cfg.Version,
	}
	return res, nil
}
