package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/services/events/protocol"
	"google.golang.org/grpc/metadata"
	"log"
)

type Server struct {
	protocol.EventsServer

	// our app context, NOT grpc
	ctx context.Context
	pub app.Pub
}

func (s *Server) Pub(ctx context.Context, req *protocol.PubReq) (*protocol.PubRes, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		meta = metadata.MD{}
	}
	log.Println(meta)

	event := &entities.Event{
		Workspace: req.Workspace,
		App:       req.App,
		Type:      req.Type,
	}
	if err := event.WithData(req.Data); err != nil {
		return nil, err
	}

	ascfg := configs.FromContext(s.ctx)
	if err := event.WithBucket(ascfg.BucketTemplate); err != nil {
		return nil, err
	}
	if err := event.WithId(); err != nil {
		return nil, err
	}

	if !event.Valid() {
		msg := fmt.Sprintf("services.events: %s is not valid event", event.Key())
		return nil, errors.New(msg)
	}

	pubkey, err := s.pub(event)
	if err != nil {
		return nil, err
	}

	res := &protocol.PubRes{
		Pubkey:     *pubkey,
		Bucket:     event.Bucket,
		Timestamps: event.Timestamps,
	}
	return res, nil
}
