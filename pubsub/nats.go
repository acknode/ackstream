package pubsub

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func NewClient(cfg *Configs) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name(cfg.Name),
	}

	return nats.Connect(cfg.Uri, opts...)
}

func NewStream(client *nats.Conn, cfg *Configs) (nats.JetStreamContext, error) {
	js, err := client.JetStream()
	if err != nil {
		return nil, err
	}

	stream, err := js.StreamInfo(cfg.Name)
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return nil, err
	}

	// stream was not initialized, we should init it
	if stream == nil {
		jsconfs := nats.StreamConfig{
			Name: cfg.Name,
			Subjects: []string{
				fmt.Sprintf("%s.%s.*", cfg.Name, cfg.Topic),
			},
			// @TODO: define MaxMsgs, MaxBytes, MaxAge, MaxMsgSize, ...
		}
		if stream, err = js.AddStream(&jsconfs); err != nil {
			return nil, err
		}
	}

	return js, nil
}

func NewPub(jsc nats.JetStreamContext, cfg *Configs) Pub {
	return func(topic string, msg *Message) (string, error) {
		natmsg := nats.NewMsg(topic)
		natmsg.Data = msg.Data

		// nats headers
		natmsg.Header.Set("Nats-Msg-Id", msg.Id)
		// copy meta to headers
		for k, v := range msg.Meta {
			natmsg.Header.Set(k, v)
		}

		ack, err := jsc.PublishMsg(natmsg)
		if err != nil {
			return "", err
		}

		key := strings.Join([]string{ack.Domain, ack.Stream, fmt.Sprint(ack.Sequence)}, "/")
		return key, nil
	}
}

func NewSub(jsc nats.JetStreamContext, cfg *Configs) Sub {
	return func(topic string, fn SubscribeFn) (func() error, error) {

		sub, err := jsc.Subscribe(topic, func(natmsg *nats.Msg) {
			msg := Message{
				Id:   natmsg.Header.Get("Nats-Msg-Id"),
				Data: natmsg.Data,
				Meta: map[string]string{},
			}
			for k, v := range natmsg.Header {
				// get first item only because we only set one value to header key
				msg.Meta[k] = v[0]
			}

			if err := fn(&msg); err != nil {
				retry := msg.GetRetryCount()
				natmsg.Header.Set(METAKEY_RETRY_COUNT, fmt.Sprint(retry+1))
				// subcribers must handle error by themself
				// if they throw an error, message will be delivered again
				natmsg.NakWithDelay(time.Duration(math.Pow(2, float64(retry))))
				return
			}

			natmsg.Ack()
		})

		// return callback to cleanup resources
		return func() error { return sub.Drain() }, err
	}
}
