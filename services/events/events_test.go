package events_test

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xrpc"
	"github.com/acknode/ackstream/services/events"
	"github.com/acknode/ackstream/services/events/protos"
	"github.com/acknode/ackstream/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"os"
	"strconv"
	"testing"
	"time"
)

func BenchmarkTestEventsPub(b *testing.B) {
	ctx := context.Background()

	ctx, err := WithLogger(ctx)
	assert.Nil(b, err)

	conn, err := xrpc.NewClient(ctx)
	assert.Nil(b, err)
	defer func() {
		_ = conn.Close()
	}()
	client, err := events.NewClient(ctx, conn)
	assert.Nil(b, err)

	meta := metadata.New(map[string]string{
		"content-type":         "application/grpc",
		"acknode-service-name": "ackstream-events",
	})
	ctx = metadata.NewOutgoingContext(ctx, meta)

	count, _ := strconv.Atoi(os.Getenv("BENCH_PARALLEL"))
	b.ResetTimer()
	if count > 0 {
		b.SetParallelism(count)
	}
	b.RunParallel(func(pb *testing.PB) {
		event := &entities.Event{
			Workspace: utils.NewId("ws"),
			App:       utils.NewId("app"),
			Type:      utils.NewId("type"),
		}
		for pb.Next() {
			ts := time.Now().UnixNano()
			req := &protos.PubReq{
				Workspace: event.Workspace,
				App:       event.App,
				Type:      event.Type,
				Data:      fmt.Sprintf(`{"ts": %d}`, ts),
			}
			if _, err := client.Pub(ctx, req); err != nil {
				b.Log(err)
			}
		}
	})
}

func WithLogger(ctx context.Context) (context.Context, error) {
	logger := xlogger.New(false)
	return xlogger.WithContext(ctx, logger), nil
}
