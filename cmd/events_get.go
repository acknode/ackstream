package cmd

import (
	"context"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
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

			session, err := xstorage.New(ctx, cfg.XStorage)
			if err != nil {
				logger.Fatal(err.Error())
			}
			defer session.Close()

			get, err := xstorage.UseGet(ctx, cfg.XStorage, session)
			if err != nil {
				logger.Fatal(err.Error())
			}

			sample := entities.Event{
				Bucket:    args[0],
				Workspace: args[1],
				App:       args[2],
				Type:      args[3],
				Id:        args[4],
			}
			e, err := get(&sample)
			if err != nil {
				logger.Fatal(err.Error())
			}

			nowrapping, _ := cmd.Flags().GetBool("nowrapping")
			draw(e, nowrapping)
		},
	}

	command.Flags().BoolP("nowrapping", "", false, "disable wrapping (or) row/column width restrictions")

	return command
}
