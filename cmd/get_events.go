package cmd

import (
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/spf13/cobra"
)

func NewGetEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		Short:             "get events",
		Example:           "ackstream get events -b 20220202 -w ws_default -a app_demo -t cli.trigger -i event_2GGUfVU4bQUcJOvBpNA9AJzQ2zI",
		PersistentPreRunE: Chain(),
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "get.events")

			ctx, err := app.Connect(cmd.Context())
			if err != nil {
				logger.Fatal(err)
			}
			defer func() {
				if _, err := app.Disconnect(ctx); err != nil {
					logger.Error(err)
				}
			}()

			sample := parseEventSample(cmd.Flags())
			if err != nil {
				logger.Fatal(err)
			}
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				logger.Fatal(err)
			}
			sample.Id = id
			if !sample.Valid() {
				logger.Fatalw("invalid event query", "key", sample.Key())
			}

			get, err := xstorage.NewGet(ctx)
			if err != nil {
				logger.Fatal(err)
			}
			event, err := get(sample)
			if err != nil {
				logger.Fatal(err)
			}

			printEvent(event)
		},
	}

	command.Flags().StringP("bucket", "b", "", " --bucket='': specify which bucket you want to scan events")
	_ = command.MarkFlagRequired("bucket")

	command.Flags().StringP("workspace", "w", "", " --workspace='': specify which workspace you want to publish an event to")
	_ = command.MarkFlagRequired("workspace")

	command.Flags().StringP("app", "a", "", "--app='': specify which app you are working with")
	_ = command.MarkFlagRequired("app")

	command.Flags().StringP("type", "t", "", "--type='': specify which type of event you want to use")
	_ = command.MarkFlagRequired("type")

	command.Flags().StringP("id", "i", "", "--id='': specify id of event you want to query")

	return command
}
