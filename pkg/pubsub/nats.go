package pubsub

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
)

func NewConn(cfg *Configs, name string) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name(fmt.Sprintf("%s-%s", name, cfg.StreamName)),
	}

	return nats.Connect(cfg.Uri, opts...)
}

func NewStream(client *nats.Conn, cfg *Configs) (nats.JetStreamContext, error) {
	js, err := client.JetStream()
	if err != nil {
		return nil, err
	}
	subject := fmt.Sprintf("%s.>", cfg.StreamName)
	stream, err := js.StreamInfo(cfg.StreamName)

	// update
	if err == nil {
		stream.Config.Subjects = lo.Uniq(append(stream.Config.Subjects, subject))
		if stream, err = js.UpdateStream(&stream.Config); err != nil {
			return nil, err
		}
	}

	if err != nil && errors.Is(err, nats.ErrStreamNotFound) {
		jscfg := nats.StreamConfig{
			Name:     cfg.StreamName,
			Subjects: []string{subject},
			// @TODO: define MaxMsgs, MaxBytes, MaxAge, MaxMsgSize, ...
		}
		if stream, err = js.AddStream(&jscfg); err != nil {
			return nil, err
		}
	}

	if stream == nil {
		return nil, err
	}

	return js, err
}

func NewPub(jsc nats.JetStreamContext, cfg *Configs) Pub {
	return func(topic string, msg *Message) (string, error) {
		// @TODO: validate topic
		natmsg := nats.NewMsg(NewSubjectFromMessage(cfg, topic, msg))
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

		key := strings.Join([]string{ack.Domain, ack.Stream, fmt.Sprint(ack.Sequence), msg.Id}, "/")
		return key, nil
	}
}

func NewSub(jsc nats.JetStreamContext, cfg *Configs) Sub {
	return func(topic, queue string, fn SubscribeFn) (func() error, error) {
		// @TODO: validate topic & queue

		subject := NewSubjectFromMessage(cfg, topic, nil)
		sub, err := jsc.QueueSubscribe(subject, queue, UseSub(fn))

		// return callback to cleanup resources
		return func() error {
			log.Println("fuck")
			return sub.Drain()
		}, err
	}
}

func UseSub(fn SubscribeFn) nats.MsgHandler {
	return func(natmsg *nats.Msg) {
		msg := Message{
			Workspace: natmsg.Header.Get(METAKEY_WORKSPACE),
			App:       natmsg.Header.Get(METAKEY_APP),
			Id:        natmsg.Header.Get("Nats-Msg-Id"),
			Data:      natmsg.Data,
			Meta:      map[string]string{},
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
	}

}
