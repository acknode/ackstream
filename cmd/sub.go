package cmd

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewSub() *cobra.Command {
	command := &cobra.Command{
		Use:               "sub -w WORKSPACE -a APP -t TYPE -q QUEUE_NAME",
		Short:             "subscribe event on stream",
		Example:           "ackstream sub -w ws_default -a app_demo -t cli.trigger -q local",
		PersistentPreRunE: Chain(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			migrateDirs, err := cmd.Flags().GetStringArray("migrate-dirs")
			if err != nil {
				return err
			}
			return xstorage.Migrate(cmd.Context(), migrateDirs)
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "sub")

			sample := parseEventSample(cmd.Flags())
			queue, err := cmd.Flags().GetString("queue")
			if err != nil {
				logger.Fatal(err)
			}

			ctx, err := app.Connect(cmd.Context())
			if err != nil {
				logger.Fatal(err)
			}
			defer func() {
				if _, err := app.Disconnect(ctx); err != nil {
					logger.Error(err.Error())
				}
			}()

			sub, err := app.NewSub(ctx)
			if err != nil {
				logger.Fatal(err)
			}

			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			logger.Info("starting...")
			go func() {
				err = sub(sample, queue, func(event *entities.Event) error {
					printEvent(event)
					return nil
				})
				if err != nil {
					logger.Fatal(err)
				}
			}()

			// Listen for the interrupt signal.
			<-ctx.Done()
			stop()
			logger.Info("shutting down gracefully, press Ctrl+C again to force")
			// The context is used to inform the server it has 5 seconds to finish
			// the request it is currently handling
			ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
			defer cancel()

			go func() {
				_, _ = app.Disconnect(ctx)
				<-ctx.Done()
			}()
		},
	}

	command.Flags().StringP("workspace", "w", "", " --workspace='': specify which workspace you want to publish an event to")
	_ = command.MarkFlagRequired("workspace")

	command.Flags().StringP("app", "a", "", "--app='': specify which app you are working with")
	_ = command.MarkFlagRequired("app")

	command.Flags().StringP("type", "t", "", "--type='': specify which type of event you want to use")
	_ = command.MarkFlagRequired("type")

	command.Flags().StringP("queue", "q", "", " --queue='': specify name of your queue. DO NOT USE PRODUCTION QUEUE NAME")
	_ = command.MarkFlagRequired("queue")

	command.Flags().StringArrayP("migrate-dirs", "", []string{}, "migrate resources before start the command")

	return command
}
