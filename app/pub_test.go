package app_test

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"testing"
	"time"
)

func BenchmarkTestPub(b *testing.B) {
	ctx := context.Background()

	ctx, _ = WithConfigs(ctx)
	ctx, _ = WithLogger(ctx)

	ctx, _ = app.Connect(ctx)
	defer func() {
		_, _ = app.Disconnect(ctx)
	}()
	pub, _ := app.NewPub(ctx)

	cfg := configs.FromContext(ctx)

	ts := time.Now().UnixNano()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			event := GenEvent(cfg, ts)
			_, _ = pub(event)
		}
	})
}
