package cmd

import (
	"context"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstorage"
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
			ctx := app.NewContext(context.Background(), logger, cfg)

			session := xstorage.New(ctx, cfg.XStorage)
			defer session.Close()
			ctx = xstorage.WithContext(ctx, session)
			logger.Debugw("load storage", "hosts", cfg.XStorage.Hosts, "keyspace", cfg.XStorage.Keyspace, "table", cfg.XStorage.Table)

			get := xstorage.UseGet(ctx, cfg.XStorage)
			e, err := get(args[0], args[1], args[2], args[3], args[4])
			if err != nil {
				logger.Fatal(err)
			}

			nowrapping, _ := cmd.Flags().GetBool("nowrapping")
			draw(e, nowrapping)
		},
	}

	command.Flags().BoolP("nowrapping", "w", false, "disable wrapping (or) row/column width restrictions")

	return command
}
