package app_test

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

func BenchmarkTestPub(b *testing.B) {
	ctx := context.Background()

	ctx, err := WithConfigs(ctx)
	assert.Nil(b, err)
	ctx, err = WithLogger(ctx)
	assert.Nil(b, err)

	ctx, err = app.Connect(ctx)
	assert.Nil(b, err)
	defer func() {
		_, _ = app.Disconnect(ctx)
	}()
	pub, err := app.NewPub(ctx)
	assert.Nil(b, err)

	cfg := configs.FromContext(ctx)

	b.ResetTimer()
	count, _ := strconv.Atoi(os.Getenv("BENCH_PARALLEL"))
	b.ResetTimer()
	if count > 0 {
		b.SetParallelism(count)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			event := GenEvent(cfg, time.Now().UnixNano())
			_, err = pub(event)
			assert.Nil(b, err)
		}
	})
}
