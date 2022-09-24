package storage_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/storage"
	"github.com/acknode/ackstream/utils"
	"github.com/ddosify/go-faker/faker"
	"github.com/gocql/gocql"
)

func Setup() (*storage.Configs, error) {
	cfg := storage.Configs{
		Keyspace: "ackstreams",
		Table:    "messages",
		Hosts:    []string{"127.0.0.1"},
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	cleanupql := fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", cfg.Keyspace)
	if err := session.Query(cleanupql).Exec(); err != nil {
		return nil, err
	}

	keyspaceql := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1};", cfg.Keyspace)
	if err := session.Query(keyspaceql).Exec(); err != nil {
		return nil, err
	}

	tableql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (bucket TEXT, workspace TEXT, app TEXT, type TEXT, id TEXT, payload TEXT, creation_time BIGINT, PRIMARY KEY ((bucket, workspace, app, type), id)) WITH CLUSTERING ORDER BY (id DESC);", cfg.Keyspace, cfg.Table)
	if err := session.Query(tableql).Exec(); err != nil {
		return nil, err
	}

	truncateql := fmt.Sprintf("TRUNCATE TABLE %s.%s;", cfg.Keyspace, cfg.Table)
	if err := session.Query(truncateql).Exec(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Init() (storage.Storage, func()) {
	cfg, err := Setup()
	if err != nil {
		log.Fatal(err)
	}

	storage := storage.KVStorage{Configs: cfg}
	if err := storage.Start(); err != nil {
		log.Fatal(err)
	}

	return &storage, func() {
		if err := storage.Stop(); err != nil {
			log.Fatal(err)
		}
	}
}

func Seed(storage storage.Storage, count int) []event.Event {
	var wg sync.WaitGroup
	locker := sync.RWMutex{}
	messages := []event.Event{}

	// init first message to init partition keys
	count--
	initmsg := NewEvent(128)
	if err := storage.Put(context.Background(), &initmsg); err == nil {
		messages = append(messages, initmsg)
	}

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locker.Lock()

			msg := NewEvent(128)
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

func NewEvent(length int) event.Event {
	f := faker.NewFaker()
	payload := fmt.Sprintf(`{"ip": "%s", "hash":"%s"}`, f.RandomIpv6(), f.RandomStringWithLength(length))
	return NewEventWithPayload(payload)
}

func NewEventWithPayload(payload string) event.Event {
	f := faker.NewFaker()

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
