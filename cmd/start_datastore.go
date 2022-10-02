package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/xstorage"
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
			logger := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "datastore")
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx, disconnect := app.NewContext(context.Background(), logger, cfg)
			defer disconnect()

			queue, _ := cmd.Flags().GetString("queue")
			ctx = context.WithValue(ctx, datastore.CTXKEY_QUEUE_NAME, queue)
			logger.Debugw("load queue", "queue", queue)

			session := xstorage.New(ctx, cfg.XStorage)
			defer session.Close()
			ctx = xstorage.WithContext(ctx, session)
			logger.Debugw("load storage", "hosts", cfg.XStorage.Hosts, "keyspace", cfg.XStorage.Keyspace, "table", cfg.XStorage.Table)

			cleanup, err := datastore.New(ctx)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			defer cleanup()
			logger.Debug("load completed")

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			logger.Debug("stopping")
		},
	}

	command.Flags().StringP("queue", "q", os.Getenv("ACKSTREAM_STREAM_QUEUE"), "specify your queue name, NOT use production queue name")

	return command
}
