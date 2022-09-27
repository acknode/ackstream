package cmd

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/pubsub"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewEvents() *cobra.Command {
	command := &cobra.Command{
		Use: "events",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := chain(cmd, args); err != nil {
				return err
			}

			return nil
		},
	}

	command.AddCommand(NewEventPub())

	return command
}

func NewEventPub() *cobra.Command {
	command := &cobra.Command{
		Use: "pub",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := chain(cmd, args); err != nil {
				return err
			}

			return nil
		},
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "events.publisher")

			ctx := context.Background()

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			conn, err := pubsub.NewConn(cfg.PubSub, "cli.datastore")
			if err != nil {
				panic(err)
			}
			defer conn.Close()
			ctx = pubsub.WithContext(ctx, conn)
			l.Debugw("load pubsub", "uri", cfg.PubSub.Uri, "stream_name", cfg.PubSub.StreamName, "stream_region", cfg.PubSub.StreamRegion)

			pub, err := app.NewPub(ctx)
			if err != nil {
				log.Fatal(err)
			}

			props, err := cmd.Flags().GetStringArray("props")
			if err != nil {
				log.Fatal(err)
			}

			payload := map[string]string{
				"app": "cli",
			}
			for _, arg := range props {
				kv := strings.Split(arg, "=")
				payload[kv[0]] = kv[1]
			}
			data, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			now := time.Now().UTC()
			msg, err := pubsub.NewMsgFromEvent(
				event.Event{
					Bucket:       utils.NewBucket(now),
					Workspace:    args[0],
					App:          args[1],
					Id:           utils.NewId("e"),
					Type:         event.TOPIC_EVENT_PUT,
					Payload:      string(data),
					CreationTime: now.UnixMicro(),
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			pubkey, err := pub(event.TOPIC_EVENT_PUT, msg)
			if err != nil {
				log.Fatal(err)
			}

			l.Infow("published", "publish_key", pubkey)
		},
	}

	command.Flags().StringArrayP("props", "p", []string{}, "message body properties")

	return command
}
