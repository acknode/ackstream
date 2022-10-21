package cmd

import (
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/spf13/cobra"
)

func NewPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub -w WORKSPACE -a APP -t TYPE [-p PROPERTY]",
		Short:             "publish an event to our stream",
		Example:           "ackstream pub -w ws_default -a app_demo -t cli.trigger -p env=local -p os=macos",
		PersistentPreRunE: UseChain(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			migrateDirs, err := cmd.Flags().GetStringArray("migrate-dirs")
			if err != nil {
				return err
			}
			return xstorage.Migrate(cmd.Context(), migrateDirs)
		},
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "pub")

			cfg := configs.FromContext(ctx)
			event, err := parseEvent(cmd.Flags())
			if err != nil {
				logger.Fatal(err)
			}
			if err := event.WithBucket(cfg.BucketTemplate); err != nil {
				logger.Fatal(err)
			}
			if err := event.WithId(); err != nil {
				logger.Fatal(err)
			}

			ctx, err = app.Connect(ctx)
			if err != nil {
				logger.Fatal(err)
			}
			defer func() {
				if _, err := app.Disconnect(ctx); err != nil {
					logger.Error(err)
				}
			}()

			pub, err := app.NewPub(ctx)
			if err != nil {
				logger.Fatal(err)
			}

			key, err := pub(event)
			if err != nil {
				logger.Fatal(err)
			}

			logger.Infow("published an event successfully", "pubkey", *key)
		},
	}

	command.Flags().StringP("workspace", "w", "", " --workspace='': specify which workspace you want to publish an event to")
	_ = command.MarkFlagRequired("workspace")

	command.Flags().StringP("app", "a", "", "--app='': specify which app you are working with")
	_ = command.MarkFlagRequired("app")

	command.Flags().StringP("type", "t", "", "--type='': specify which type of event you want to use")
	_ = command.MarkFlagRequired("type")

	command.Flags().StringArrayP("props", "p", []string{}, "--props='env=local': set data properties")
	command.Flags().StringArrayP("migrate-dirs", "", []string{}, "migrate resources before start the command")

	return command
}
