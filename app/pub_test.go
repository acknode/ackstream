package app_test

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"sync"
	"testing"
)

func BenchmarkPub(b *testing.B) {
	ctx := context.Background()

	ctx, _ = WithConfigs(ctx)
	ctx, _ = WithLogger(ctx)

	ctx, _ = app.Connect(ctx)
	defer func() {
		_, _ = app.Disconnect(ctx)
	}()
	pub, _ := app.NewPub(ctx)

	cfg := configs.FromContext(ctx)

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			event := GenEvent(cfg, i)
			_, _ = pub(event)
			wg.Done()
		}()
	}
	wg.Wait()
}
