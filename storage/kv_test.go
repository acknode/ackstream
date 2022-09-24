package storage_test

import (
	"context"
	"log"
	"math"
	"testing"

	"github.com/acknode/ackstream/storage"
	"github.com/ddosify/go-faker/faker"
	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	cfg, err := Setup()
	if err != nil {
		log.Fatal(err)
	}

	storage := storage.KVStorage{Configs: cfg}

	// at begining, those values are nil
	assert.Nil(t, storage.Cluster)
	assert.Nil(t, storage.Session)

	// after start we should init necessary values
	assert.NoError(t, storage.Start())
	assert.NotNil(t, storage.Cluster)
	assert.NotNil(t, storage.Session)
	assert.Equal(t, storage.Cluster.Hosts, cfg.Hosts)
	assert.Equal(t, storage.Cluster.Keyspace, cfg.Keyspace)

	// start again will not throw any error
	assert.NoError(t, storage.Start())

	// after stop all values should be set to nil
	assert.NoError(t, storage.Stop())
	assert.Nil(t, storage.Cluster)
	assert.Nil(t, storage.Session)

	// stop again will not throw any error
	assert.NoError(t, storage.Stop())
}

func TestCRUD(t *testing.T) {
	storage, cleanup := Init()
	defer cleanup()

	msg := NewEvent(512 * 1024)

	// push should be successful
	err := storage.Put(context.Background(), &msg)
	assert.Nil(t, err)

	// get inserted message
	foundmsg, err := storage.Get(context.Background(), msg.Bucket, msg.Workspace, msg.App, msg.Type, msg.Id)
	assert.Nil(t, err)

	assert.Equal(t, foundmsg.Payload, msg.Payload)

	assert.Nil(t, err)
	assert.Equal(t, msg.Payload, foundmsg.Payload)

	assert.Equal(t, foundmsg.CreationTime, msg.CreationTime)

	msgs, page, errs := storage.Scan(context.Background(), msg.Bucket, msg.Workspace, msg.App, msg.Type, 10, []byte{})
	assert.Empty(t, errs)

	assert.Equal(t, len(page), 0)
	assert.Equal(t, len(msgs), 1)
}

func TestScan(t *testing.T) {
	storage, cleanup := Init()
	defer cleanup()

	// Seed arbitrary number of messages
	f := faker.NewFaker()
	count := f.RandomIntBetween(499, 999)
	messages := Seed(storage, count)
	assert.Equal(t, count, len(messages))

	var page []byte
	size := 100
	found := 0

	// round up to nearest int
	round := int(math.Round(float64(count/size) + 0.5))

	for i := 0; i < round; i++ {
		foundmessages, newPage, errs := storage.Scan(
			context.Background(),
			messages[0].Bucket,
			messages[0].Workspace,
			messages[0].App,
			messages[0].Type,
			size, page,
		)
		assert.Empty(t, errs)

		found += len(foundmessages)
		page = newPage
	}

	assert.Equal(t, count, found)
}
