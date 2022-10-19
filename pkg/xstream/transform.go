package xstream

import (
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/nats-io/nats.go"
	"strconv"
)

func NewMsg(cfg *Configs, event *entities.Event) (*nats.Msg, error) {
	msg := nats.NewMsg(NewSubject(cfg, event))
	msg.Data = []byte(event.Data)

	// with metadata
	msg.Header.Set("Nats-Msg-Id", event.Id)
	msg.Header.Set("AckStream-Event-Id", event.Id)
	msg.Header.Set("AckStream-Event-Bucket", event.Bucket)
	msg.Header.Set("AckStream-Event-Workspace", event.Workspace)
	msg.Header.Set("AckStream-Event-App", event.App)
	msg.Header.Set("AckStream-Event-Type", event.Type)
	msg.Header.Set("AckStream-Event-Timestamps", fmt.Sprint(event.Timestamps))

	return msg, nil
}

func NewEvent(msg *nats.Msg) (*entities.Event, error) {
	event := entities.Event{
		Id:        msg.Header.Get("AckStream-Event-Id"),
		Bucket:    msg.Header.Get("AckStream-Event-Bucket"),
		Workspace: msg.Header.Get("AckStream-Event-Workspace"),
		App:       msg.Header.Get("AckStream-Event-App"),
		Type:      msg.Header.Get("AckStream-Event-Type"),
		Data:      string(msg.Data),
	}

	ts, err := strconv.ParseInt(msg.Header.Get("AckStream-Event-Timestamps"), 10, 64)
	if err != nil {
		return nil, err
	}
	event.Timestamps = ts

	if !event.Valid() {
		return nil, ErrMsgInvalidEvent
	}
	return &event, nil
}
