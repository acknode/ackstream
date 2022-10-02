package cmd

import (
	"context"
	"log"
	"strings"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/pkg/configs"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewEventsPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub",
		PersistentPreRunE: Chain(),
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

			pubkey, err := pub(args[0], args[1], args[2], data)
			if err != nil {
				log.Fatal(err)
			}

			logger.Infow("published", "publish_key", pubkey)
		},
	}

	command.Flags().StringArrayP("props", "p", []string{}, "message body properties")

	return command
}
