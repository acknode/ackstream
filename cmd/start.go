package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/logger"
	"github.com/acknode/ackstream/internal/xstorage"
	"github.com/acknode/ackstream/internal/xstream"
	"github.com/acknode/ackstream/services/datastore"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewStart() *cobra.Command {
	command := &cobra.Command{
		Use: "start",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chain := clichain()
			if err := chain(cmd, args); err != nil {
				return err
			}

			auto, err := cmd.Flags().GetBool("auto-migrate")
			if err != nil {
				return err
			}

			l := cmd.Context().Value(CTXKEY_LOGGER).(*zap.SugaredLogger)
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			if auto && !cfg.Debug {
				l.Warnw("set auto migrate but environment is not development")
			}
			// migrate storage before start
			if auto && cfg.Debug {
				l.Debugw("migrating", "hosts", cfg.Storage.Hosts, "keyspace", cfg.Storage.Keyspace, "table", cfg.Storage.Table)
				return xstorage.Migrate(cfg.Storage)
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
		Use: "datastore",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chain := clichain()
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
			l.Debugw("load queue", "queue", queue)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			session := xstorage.New(ctx, cfg.Storage)
			defer session.Close()
			ctx = xstorage.WithContext(ctx, session)
			l.Debugw("load storage", "hosts", cfg.Storage.Hosts, "keyspace", cfg.Storage.Keyspace, "table", cfg.Storage.Table)

			stream, conn := xstream.New(ctx, cfg.Stream)
			ctx = xstream.WithContext(ctx, conn, stream)
			l.Debugw("load stream", "region", cfg.Stream.Region, "uri", cfg.Stream.Uri, "name", cfg.Stream.Name)

			cleanup, err := datastore.New(ctx)
			if err != nil {
				l.Error(err.Error())
				return
			}
			defer cleanup()
			l.Info("load completed")

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			l.Info("stopping")
		},
	}

	command.Flags().StringP("queue", "q", os.Getenv("ACKSTREAM_STREAM_QUEUE"), "specify your queue name, NOT use production queue name")

	return command
}
