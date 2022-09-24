package pubsub_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/acknode/ackstream/pubsub"
	"github.com/nats-io/nats-server/v2/server"
	nattest "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestNatsPubSub(t *testing.T) {
	server, opts := NewNatsServer()
	defer server.Shutdown()

	cfg := pubsub.Configs{
		Name:  "ackstream",
		Uri:   fmt.Sprintf("nats://127.0.0.1:%d", opts.Port),
		Topic: "events",
	}

	client, err := pubsub.NewClient(&cfg)
	assert.Nil(t, err)

	jsc, err := pubsub.NewStream(client, &cfg)
	assert.Nil(t, err)

	// init publish function
	publish := pubsub.NewPub(jsc, &cfg)

	// make sure we cleanup messages before doing the test
	jsc.PurgeStream(cfg.Name)

	msg, err := pubsub.MsgFromEvent(NewEvent())
	assert.Nil(t, err)

	eventtopic := fmt.Sprintf("%s.%s.%s", cfg.Name, cfg.Topic, "put")

	pubkey, err := publish(eventtopic, msg)
	assert.Nil(t, err)
	assert.NotEmpty(t, pubkey)
	// make sure stream was stored our msg successfully
	stream, _ := jsc.StreamInfo(cfg.Name)
	assert.NotNil(t, stream)
	assert.Equal(t, stream.State.Msgs, uint64(1))

	// subscribe later
	subscribe := pubsub.NewSub(jsc, &cfg)

	var acktime int64
	cleanup, err := subscribe(eventtopic, func(natmsg *pubsub.Message) error {
		assert.Equal(t, natmsg.Id, msg.Id)
		assert.Equal(t, natmsg.Data, msg.Data)

		for k, v := range msg.Meta {
			assert.Equal(t, natmsg.Meta[k], v)
		}
		assert.NotNil(t, natmsg.Meta["Nats-Msg-Id"])
		acktime = time.Now().UnixMicro()
		return nil
	})
	assert.Nil(t, err)
	defer cleanup()

	for i := 1; i < 5; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		if acktime > 0 {
			break
		}
	}

	// and make sure message could be delivered
	assert.Greater(t, acktime, int64(0))
}

func NewNatsServer() (*server.Server, *server.Options) {
	opts := nattest.DefaultTestOptions
	opts.Port = 4242
	opts.JetStream = true
	return nattest.RunServer(&opts), &opts
}
