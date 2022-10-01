package cmd

import (
	"context"
	"time"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/xstorage"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewEventsGet() *cobra.Command {
	command := &cobra.Command{
		Use:               "get",
		PersistentPreRunE: Chain(),
		Args:              cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			logger := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "cli.events.get")
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx, disconnect := app.NewContext(context.Background(), logger, cfg)
			defer disconnect()

			session := xstorage.New(ctx, cfg.XStorage)
			defer session.Close()
			ctx = xstorage.WithContext(ctx, session)
			logger.Debugw("load storage", "hosts", cfg.XStorage.Hosts, "keyspace", cfg.XStorage.Keyspace, "table", cfg.XStorage.Table)

			get := xstorage.UseGet(ctx, cfg.XStorage)
			e, err := get(args[0], args[1], args[2], args[3], args[4])
			if err != nil {
				logger.Fatal(err)
			}

			logger.Infow("got event",
				"key", e.Key(),
				"data", e.Data,
				"creation_time", time.UnixMicro(e.CreationTime).Format(time.RFC3339),
			)
		},
	}

	return command
}
