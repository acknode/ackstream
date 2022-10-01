package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewEventsSub() *cobra.Command {
	command := &cobra.Command{
		Use:               "sub",
		PersistentPreRunE: Chain(),
		Args:              cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			queue := args[0]

			logger := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("queue", queue).
				With("service", "cli.events.subscriber")
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx, disconnect := app.NewContext(context.Background(), logger, cfg)
			defer disconnect()

			cb, err := app.UseSub(ctx, queue, func(e *event.Event) error {
				log.Printf("event: bucket=%s ws=%s app=%s type=%s id=%s", e.Bucket, e.Workspace, e.App, e.Type, e.Id)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
			defer cb()

			logger.Info("subscribing")
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit
		},
	}

	command.Flags().StringArrayP("props", "p", []string{}, "message body properties")

	return command
}
