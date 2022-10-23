package app_test

import (
	"context"
	"fmt"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/utils"
	"os"
	"path"
)

func WithConfigs(ctx context.Context) (context.Context, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return ctx, err
	}

	provider, err := configs.NewProvider(cwd, path.Join(cwd, "secrets"))
	if err != nil {
		return ctx, err
	}

	cfg, err := configs.New(provider, []string{"ACKSTREAM_ENV=dev"})
	if err != nil {
		return ctx, err
	}

	return configs.WithContext(ctx, cfg), nil
}

func WithLogger(ctx context.Context) (context.Context, error) {
	cfg := configs.FromContext(ctx)
	logger := xlogger.New(cfg.Debug)
	return xlogger.WithContext(ctx, logger), nil
}

func GenEvent(cfg *configs.Configs, i int) *entities.Event {
	event := &entities.Event{
		Workspace: utils.NewId("ws"),
		App:       utils.NewId("app"),
		Type:      utils.NewId("type"),
		Data:      fmt.Sprintf(`{"i": %d}`, i),
	}
	_ = event.WithId()
	_ = event.WithBucket(cfg.BucketTemplate)
	return event
}
