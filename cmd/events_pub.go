package cmd

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewEventsPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub",
		PersistentPreRunE: clichain(),
		Args:              cobra.ExactArgs(3),
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

			data := map[string]string{"app": "cli"}
			for _, arg := range props {
				kv := strings.Split(arg, "=")
				data[kv[0]] = kv[1]
			}

			now := time.Now().UTC()
			e := event.Event{
				Bucket:       utils.NewBucket(now),
				Workspace:    args[0],
				App:          args[1],
				Type:         args[2],
				Id:           utils.NewId("e"),
				Data:         data,
				CreationTime: now.UnixMicro(),
			}
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
