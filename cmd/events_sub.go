package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/pkg/xstream"
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
			ctx := app.NewContext(context.Background(), logger, cfg)

			ctx, err := xstream.Connect(ctx, cfg.XStream)
			if err != nil {
				logger.Fatal(err)
			}

			sub, err := xstream.NewSub(ctx, cfg.XStream)
			if err != nil {
				logger.Fatal(err.Error())
			}

			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			go func() {
				nowrapping, _ := cmd.Flags().GetBool("nowrapping")
				ctx, err = sub(getSampleEvent(cmd.Flags(), false), queue, func(e *entities.Event) error {
					draw(e, nowrapping)
					return nil
				})
				if err != nil {
					logger.Fatal(err.Error())
				}

				logger.Info("subscribing")
			}()

			// Listen for the interrupt signal.
			<-ctx.Done()
			stop()
			logger.Info("shutting down gracefully, press Ctrl+C again to force")
			// The context is used to inform the server it has 5 seconds to finish
			// the request it is currently handling
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := xstream.Disconnect(ctx, cfg.XStream); err != nil {
				logger.Fatal(err.Error())
			}
		},
	}

	command.Flags().StringP("queue", "q", os.Getenv("ACKSTREAM_STREAM_QUEUE"), "specify your queue name, NOT use production queue name")
	command.Flags().BoolP("nowrapping", "", false, "disable wrapping (or) row/column width restrictions")

	return command
}
