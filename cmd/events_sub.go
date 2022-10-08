package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewEventsSub() *cobra.Command {
	command := &cobra.Command{
		Use: "sub",
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
			queue, _ := cmd.Flags().GetString("queue")

			logger := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("queue", queue).
				With("service", "cli.events.subscriber")
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx, disconnect := app.NewContext(context.Background(), logger, cfg)
			defer disconnect()

			nowrapping, _ := cmd.Flags().GetBool("nowrapping")
			cb, err := app.UseSub(ctx, getSampleEvent(cmd.Flags(), false), queue, func(e *entities.Event) error {
				draw(e, nowrapping)
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

	command.Flags().StringP("queue", "q", os.Getenv("ACKSTREAM_STREAM_QUEUE"), "specify your queue name, NOT use production queue name")
	command.Flags().BoolP("nowrapping", "", false, "disable wrapping (or) row/column width restrictions")

	return command
}
