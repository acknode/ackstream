package events_test

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events"
	eventscfg "github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/protos"
	"github.com/acknode/ackstream/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

func BenchmarkTestEventsPub(b *testing.B) {
	ctx := context.Background()

	ctx, err := WithConfigs(ctx)
	assert.Nil(b, err)
	ctx, err = WithLogger(ctx)
	assert.Nil(b, err)

	conn, client, _ := events.NewClient(ctx)
	defer func() {
		_ = conn.Close()
	}()

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

func WithConfigs(ctx context.Context) (context.Context, error) {
	cwd, err := utils.GetRootDir()
	if err != nil {
		return ctx, err
	}
	provider, err := eventscfg.NewProvider(*cwd, path.Join(*cwd, "secrets"))
	if err != nil {
		return ctx, err
	}

	cfg, err := eventscfg.New(provider, []string{"ACKSTREAM_ENV=test"})
	if err != nil {
		return ctx, err
	}

	return eventscfg.WithContext(ctx, cfg), nil
}

func WithLogger(ctx context.Context) (context.Context, error) {
	logger := xlogger.New(false)
	return xlogger.WithContext(ctx, logger), nil
}
