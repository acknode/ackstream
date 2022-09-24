package pubsub

import (
	"context"
	"errors"
	"strconv"

	"github.com/acknode/ackstream/event"
	"github.com/vmihailenco/msgpack/v5"
)

const MSG_MAX_SIZE = 1572864

type Message struct {
	Id   string
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

type Sub func(topic string, fn SubscribeFn) (func() error, error)

type Pub func(topic string, msg *Message) (string, error)

type Configs struct {
	Uri   string `json:"uri" mapstructure:"ACKSTREAM_PUBSUB_URI"`
	Name  string `json:"name" mapstructure:"ACKSTREAM_PUBSUB_NAME"`
	Topic string `json:"stream_topics" mapstructure:"ACKSTREAM_PUBSUB_TOPIC"`
}

func MsgFromEvent(e event.Event) (*Message, error) {
	data, err := msgpack.Marshal(e)
	if err != nil {
		return nil, err
	}

	return &Message{Id: e.Id, Data: data, Meta: map[string]string{}}, nil
}

const (
	CTXKEY_CLIENT       string = "ackstream.pubsub.client"
	METAKEY_RETRY_COUNT string = "AckStream-Retry-Count"
)

func FromContext[C any](ctx context.Context) (*C, error) {
	client, ok := ctx.Value(CTXKEY_CLIENT).(*C)
	if !ok {
		return nil, errors.New("no pubsub client was configured")
	}

	return client, nil
}
