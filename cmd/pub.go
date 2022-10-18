package cmd

import (
	"encoding/json"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/spf13/cobra"
	"strings"
)

func NewPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub -w WORKSPACE -a APP -t TYPE [-p PROPERTY]",
		Short:             "publish an event to our stream",
		Example:           "ackstream pub -w ws_default -a app_demo -t cli.trigger -p env=local -p os=macos",
		PersistentPreRunE: Chain(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			migrateDirs, err := cmd.Flags().GetStringArray("migrate-dirs")
			if err != nil {
				return err
			}
			return xstorage.Migrate(cmd.Context(), migrateDirs)
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "pub")

			event := parseEventSample(cmd.Flags())
			cfg := configs.FromContext(cmd.Context())
			if err := event.WithBucket(cfg.BucketTemplate); err != nil {
				logger.Fatal(err)
			}
			if err := event.WithId(); err != nil {
				logger.Fatal(err)
			}

			data := map[string]interface{}{}
			props, err := cmd.Flags().GetStringArray("props")
			if err != nil {
				logger.Fatal(err)
			}
			for _, prop := range props {
				kv := strings.Split(prop, "=")
				data[kv[0]] = kv[1]
			}
			bytes, err := json.Marshal(data)
			if err != nil {
				logger.Fatal(err)
			}
			event.Data = string(bytes)

			ctx, err := app.Connect(cmd.Context())
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
