package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/services/datastore"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewStartDatastore() *cobra.Command {
	command := &cobra.Command{
		Use: "datastore",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chain := Chain()
			if err := chain(cmd, args); err != nil {
				return err
			}

			queue, err := cmd.Flags().GetString("queue")
			if err != nil || queue == "" {
				return errors.New("no queue name was configured")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger := cmd.Context().Value(CTXKEY_LOGGER).(*zap.SugaredLogger)
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx := app.NewContext(context.Background(), logger, cfg)

			queue, _ := cmd.Flags().GetString("queue")
			ctx = context.WithValue(ctx, datastore.CTXKEY_QUEUE_NAME, queue)
			logger.Debugw("load queue", "queue", queue)

			logger.Debug("load completed")
			if err := datastore.New(ctx, cfg); err != nil {
				logger.Fatal(err.Error())
			}

			logger.Debug("stopping")
		},
	}

	command.Flags().StringP("queue", "q", os.Getenv("ACKSTREAM_STREAM_QUEUE"), "specify your queue name, NOT use production queue name")

	return command
}
