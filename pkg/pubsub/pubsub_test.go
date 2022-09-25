package pubsub_test

import (
	"fmt"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/utils"
	"github.com/ddosify/go-faker/faker"
)

func NewEvent() event.Event {
	f := faker.NewFaker()

	return event.Event{
		Bucket:       utils.NewBucket(time.Now().UTC()),
		Workspace:    f.RandomUUID().String(),
		App:          f.RandomUUID().String(),
		Type:         f.RandomBsNoun(),
		Id:           utils.NewId("event"),
		Payload:      fmt.Sprintf(`{"ip": "%s", "hash":"%s"}`, f.RandomIpv6(), f.RandomStringWithLength(512*1024)),
		CreationTime: time.Now().UTC().UnixMilli(),
	}
}
