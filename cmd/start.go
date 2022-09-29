package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/logger"
	"github.com/acknode/ackstream/internal/storage"
	"github.com/acknode/ackstream/services/datastore"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewStart() *cobra.Command {
	command := &cobra.Command{
		Use: "start",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chain := useChain()
			if err := chain(cmd, args); err != nil {
				return err
			}

			auto, err := cmd.Flags().GetBool("auto-migrate")
			if err != nil {
				return err
			}

			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger)

			// migrate storage before start
			if auto {
				cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
				l.Debugw("migrating", "hosts", cfg.Storage.Hosts, "keyspace", cfg.Storage.Keyspace, "table", cfg.Storage.Table)
				return storage.Migrate(cfg.Storage)
			}

			return nil
		},
	}

	command.AddCommand(NewStartDatastore())

	command.PersistentFlags().BoolP("auto-migrate", "", false, "run auto-migration process when start an application")

	return command
}

func NewStartDatastore() *cobra.Command {
	command := &cobra.Command{
		Use:               "datastore",
		PersistentPreRunE: useChain(),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "datastore")
			ctx = logger.WithContext(ctx, l)

			queue, err := cmd.Flags().GetString("queue")
			if err != nil {
				l.Error("no queue name was configured")
				return
			}

			ctx = context.WithValue(ctx, datastore.CTXKEY_QUEUE_NAME, queue)
			l.Debugw("load queue name", "queue", queue)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			client := storage.New(cfg.Storage)
			if err := client.Start(); err != nil {
				panic(err)
			}
			defer client.Stop()
			ctx = storage.WithContext(ctx, client)
			l.Debugw("load storage", "hosts", cfg.Storage.Hosts, "keyspace", cfg.Storage.Keyspace, "table", cfg.Storage.Table)

			cleanup, err := datastore.New(ctx)
			if err != nil {
				panic(err)
			}
			defer cleanup()
			l.Info("load completed")

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			l.Info("stopping")
		},
	}

	command.Flags().StringP("queue", "q", "cli", "specify your queue name, NOT use production queue name")
	command.MarkFlagRequired("queue")

	return command
}
