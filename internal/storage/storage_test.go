package storage_test

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/storage"
	"github.com/acknode/ackstream/utils"
	"github.com/ddosify/go-faker/faker"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	storage, cleanup := setup()
	defer cleanup()

	assert.NotNil(t, storage.Cluster)
	assert.NotNil(t, storage.Session)
	assert.Equal(t, storage.Cluster.Hosts, storage.Configs.Hosts)
	assert.Equal(t, storage.Cluster.Keyspace, storage.Configs.Keyspace)

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
	storage, cleanup := setup()
	defer cleanup()

	msg := genEvent(512 * 1024)

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
	storage, cleanup := setup()
	defer cleanup()

	// seed arbitrary number of messages
	f := faker.NewFaker()
	count := f.RandomIntBetween(499, 999)
	messages := seed(storage, count)
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

func genEvent(length int) event.Event {
	f := faker.NewFaker()
	payload := fmt.Sprintf(`{"ip": "%s", "hash":"%s"}`, f.RandomIpv6(), f.RandomStringWithLength(length))

	return event.Event{
		Bucket:       utils.NewBucket(time.Now().UTC()),
		Workspace:    f.RandomUUID().String(),
		App:          f.RandomUUID().String(),
		Type:         f.RandomBsNoun(),
		Id:           utils.NewId("msg"),
		Payload:      payload,
		CreationTime: time.Now().UTC().UnixMilli(),
	}
}

func setup() (*storage.Storage, func()) {
	cfg := storage.Configs{
		Keyspace: "ackstream",
		Table:    "messages",
		Hosts:    []string{"127.0.0.1"},
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	cleanupql := fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", cfg.Keyspace)
	if err := session.Query(cleanupql).Exec(); err != nil {
		log.Fatal(err)
	}

	if err := storage.Migrate(&cfg); err != nil {
		log.Fatal(err)
	}

	truncateql := fmt.Sprintf("TRUNCATE TABLE %s.%s;", cfg.Keyspace, cfg.Table)
	if err := session.Query(truncateql).Exec(); err != nil {
		log.Fatal(err)
	}

	storage := storage.Storage{Configs: &cfg}
	if err := storage.Start(); err != nil {
		log.Fatal(err)
	}

	return &storage, func() {
		if err := storage.Stop(); err != nil {
			log.Fatal(err)
		}
	}
}

func seed(storage *storage.Storage, count int) []event.Event {
	var wg sync.WaitGroup
	locker := sync.RWMutex{}
	messages := []event.Event{}

	// init first message to init partition keys
	count--
	initmsg := genEvent(128)
	if err := storage.Put(context.Background(), &initmsg); err == nil {
		messages = append(messages, initmsg)
	}

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locker.Lock()

			msg := genEvent(128)
			msg.SetPartitionKeys(&initmsg)

			if err := storage.Put(context.Background(), &msg); err == nil {
				messages = append(messages, msg)
			}

			locker.Unlock()
		}()
	}

	wg.Wait()
	return messages
}
