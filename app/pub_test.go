package app_test

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/stretchr/testify/assert"
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

	ts := time.Now().UnixNano()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			event := GenEvent(cfg, ts)
			_, err = pub(event)
			assert.Nil(b, err)
		}
	})
}
