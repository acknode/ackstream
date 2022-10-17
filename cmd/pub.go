package cmd

import (
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/cobra"
	"strings"
)

func NewPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub -w workspace -a app -t type [-p property]",
		Short:             "publish an event to our stream",
		Example:           "ackstream pub -w ws_default -a app_demo -t cli.trigger -p env=local -p os=macos",
		PersistentPreRunE: Chain(),
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "pub")

			ws, err := cmd.Flags().GetString("workspace")
			if err != nil {
				logger.Fatal(err.Error())
			}
			eapp, err := cmd.Flags().GetString("app")
			if err != nil {
				logger.Fatal(err.Error())
			}
			etype, err := cmd.Flags().GetString("type")
			if err != nil {
				logger.Fatal(err.Error())
			}

			cfg := configs.FromContext(cmd.Context())
			bucket, ts := utils.NewBucket(cfg.BucketTemplate)
			event := &entities.Event{
				Bucket:     bucket,
				Workspace:  ws,
				App:        eapp,
				Type:       etype,
				Id:         utils.NewId("event"),
				Timestamps: ts,
				Data:       map[string]interface{}{},
			}

			props, err := cmd.Flags().GetStringArray("props")
			if err != nil {
				logger.Fatal(err.Error())
			}
			for _, prop := range props {
				kv := strings.Split(prop, "=")
				event.Data[kv[0]] = kv[1]
			}

			ctx, err := app.Connect(cmd.Context())
			if err != nil {
				logger.Fatal(err.Error())
			}
			defer func() {
				if _, err := app.Disconnect(ctx); err != nil {
					logger.Error(err.Error())
				}
			}()

			pub, err := app.NewPub(ctx)
			if err != nil {
				logger.Fatal(err.Error())
			}

			key, err := pub(event)
			if err != nil {
				logger.Fatal(err.Error())
			}

			logger.Info("published an event successfully", "pubkey", key)
		},
	}

	command.Flags().StringP("workspace", "w", "", " --workspace='': specify which workspace you want to publish an event to")
	_ = command.MarkFlagRequired("workspace")

	command.Flags().StringP("app", "a", "", "--app='': specify which app you are working with")
	_ = command.MarkFlagRequired("app")

	command.Flags().StringP("type", "t", "", "--type='': specify which type of event you want to use")
	_ = command.MarkFlagRequired("type")

	command.Flags().StringArrayP("props", "p", []string{}, "--props='env=local': set data properties")

	return command
}
