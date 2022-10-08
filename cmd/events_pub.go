package cmd

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/entities"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewEventsPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub",
		PersistentPreRunE: Chain(),
		Run: func(cmd *cobra.Command, args []string) {
			logger := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "cli.events.publisher")
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx, disconnect := app.NewContext(context.Background(), logger, cfg)
			defer disconnect()

			pub := app.UsePub(ctx)
			props, err := cmd.Flags().GetStringArray("props")
			if err != nil {
				log.Fatal(err)
			}

			maps := map[string]string{"app": "cli"}
			for _, arg := range props {
				kv := strings.Split(arg, "=")
				maps[kv[0]] = kv[1]
			}
			data, err := json.Marshal(maps)
			if err != nil {
				log.Fatal(err)
			}

			sample := getSampleEvent(cmd.Flags(), true)
			bucket, ts := utils.NewBucket(cfg.XStorage.BucketTemplate)
			e := entities.Event{
				Bucket:       bucket,
				Workspace:    sample.Workspace,
				App:          sample.App,
				Type:         sample.Type,
				CreationTime: ts,
				Data:         string(data),
			}
			e.WithId()

			pubkey, err := pub(&e)
			if err != nil {
				log.Fatal(err)
			}

			logger.Infow("published", "publish_key", pubkey)
		},
	}

	command.Flags().StringArrayP("props", "p", []string{}, "message body properties")

	return command
}
