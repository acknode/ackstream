package pubsub_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/acknode/ackstream/pkg/pubsub"
	"github.com/nats-io/nats-server/v2/server"
	natstest "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestNatsPubSub(t *testing.T) {
	server, opts := NewNatsServer()
	defer server.Shutdown()

	TOPIC := "events.put"
	QUEUE := "stdout"
	cfg := pubsub.Configs{
		Uri:          fmt.Sprintf("nats://127.0.0.1:%d", opts.Port),
		StreamRegion: "local",
		StreamName:   "ackstream",
	}

	client, err := pubsub.NewConn(&cfg, "testing")
	assert.Nil(t, err)

	jsc, err := pubsub.NewStream(client, &cfg)
	assert.Nil(t, err)

	// make sure we cleanup messages before doing the test
	err = jsc.PurgeStream(cfg.StreamName)
	assert.Nil(t, err)

	// init publish function
	publish := pubsub.NewPub(jsc, &cfg)

	msg, err := pubsub.NewMsgFromEvent(NewEvent())
	assert.Nil(t, err)

	// publish first
	pubkey, err := publish(TOPIC, msg)
	assert.Nil(t, err)
	assert.NotEmpty(t, pubkey)
	// make sure stream was stored our msg successfully
	stream, _ := jsc.StreamInfo(cfg.StreamName)
	assert.NotNil(t, stream)
	assert.Equal(t, stream.State.Msgs, uint64(1))

	// Make duplicated push
	_, err = publish(TOPIC, msg)
	assert.Nil(t, err)

	// subscribe later
	subscribe := pubsub.NewSub(jsc, &cfg)

	var acktime int64
	cleanup, err := subscribe(TOPIC, QUEUE, func(natmsg *pubsub.Message) error {
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

	for i := 1; i < 10; i++ {
		time.Sleep(time.Duration(i*100) * time.Millisecond)
		if acktime > 0 {
			break
		}
	}

	// and make sure message could be delivered
	assert.Greater(t, acktime, int64(0))
}

func NewNatsServer() (*server.Server, *server.Options) {
	opts := natstest.DefaultTestOptions
	opts.Port = 4242
	opts.JetStream = true
	return natstest.RunServer(&opts), &opts
}
