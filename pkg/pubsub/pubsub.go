package pubsub

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/acknode/ackstream/event"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

const MSG_MAX_SIZE = 1572864

type Message struct {
	Workspace string
	App       string
	Id        string

	Data []byte
	Meta map[string]string
}

func (msg *Message) GetRetryCount() int {
	retry, ok := msg.Meta[METAKEY_RETRY_COUNT]
	if !ok {
		return 0
	}

	count, err := strconv.Atoi(retry)
	if err != nil {
		return 0
	}

	return count
}

type SubscribeFn func(msg *Message) error

type Sub func(topic, queue string, fn SubscribeFn) (func() error, error)

type Pub func(topic string, msg *Message) (string, error)

type Configs struct {
	Uri          string `json:"uri" mapstructure:"ACKSTREAM_PUBSUB_URI"`
	StreamRegion string `json:"region" mapstructure:"ACKSTREAM_PUBSUB_STREAM_REGION"`
	StreamName   string `json:"name" mapstructure:"ACKSTREAM_PUBSUB_STREAM_NAME"`
}

type ctxkey string

const (
	CTXKEY_CONN         ctxkey = "ackstream.pubsub.conn"
	METAKEY_WORKSPACE   string = "AckStream-Workspace"
	METAKEY_APP         string = "AckStream-App"
	METAKEY_RETRY_COUNT string = "AckStream-Retry-Count"
)

func WithContext(ctx context.Context, conn *nats.Conn) context.Context {
	return context.WithValue(ctx, CTXKEY_CONN, conn)
}

func FromContext(ctx context.Context) (*nats.Conn, error) {
	conn, ok := ctx.Value(CTXKEY_CONN).(*nats.Conn)
	if !ok {
		return nil, errors.New("no pubsub connection was configured")
	}

	return conn, nil
}

func NewSubjectFromMessage(cfg *Configs, topic string, msg *Message) string {
	// using wildcard if msg is nil
	if msg == nil {
		return strings.Join([]string{cfg.StreamName, topic, ">"}, ".")
	}
	return strings.Join([]string{cfg.StreamName, topic, msg.Workspace, msg.App}, ".")
}

func NewMsgFromEvent(e event.Event) (*Message, error) {
	data, err := msgpack.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Workspace: e.Workspace,
		App:       e.App,
		Id:        e.Id,
		Data:      data,
		Meta: map[string]string{
			METAKEY_WORKSPACE: e.Workspace,
			METAKEY_APP:       e.App,
		},
	}
	return &msg, nil
}
