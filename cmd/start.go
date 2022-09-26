package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/storage"
	"github.com/acknode/ackstream/pkg/pubsub"
	"github.com/acknode/ackstream/services/datastore"
	"github.com/spf13/cobra"
)

func NewStart() *cobra.Command {
	command := &cobra.Command{
		Use: "start",
	}

	command.AddCommand(NewStartDatastore())

	return command
}

func NewStartDatastore() *cobra.Command {
	command := &cobra.Command{
		Use: "datastore",
		Run: func(cmd *cobra.Command, args []string) {
			queue, err := cmd.Flags().GetString("queue")
			if err != nil {
				panic(err)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, datastore.CTXKEY_QUEUE_NAME, queue)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)

			conn, err := pubsub.NewConn(cfg.PubSub, "cli.datastore")
			if err != nil {
				panic(err)
			}
			defer conn.Close()
			ctx = pubsub.WithContext(ctx, conn)

			client := storage.New(cfg.Storage)
			if err := client.Start(); err != nil {
				panic(err)
			}
			defer client.Stop()
			ctx = storage.WithContext(ctx, client)

			run, err := datastore.New(ctx)
			if err != nil {
				panic(err)
			}

			go func() {
				if err := run(); err != nil {
					panic(err)
				}
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit
		},
	}

	command.Flags().StringP("queue", "q", "cli", "specify your queue name, NOT use production queue name")
	command.MarkFlagRequired("queue")

	return command
}
