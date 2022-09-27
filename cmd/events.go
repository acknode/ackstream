package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/logger"
	"github.com/acknode/ackstream/pkg/pubsub"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/cobra"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

func NewEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		PersistentPreRunE: useChain(),
	}

	command.AddCommand(NewEventsPub())
	command.AddCommand(NewEventsSub())

	return command
}

func NewEventsPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub",
		PersistentPreRunE: useChain(),
		Args:              cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "events.publisher")
			ctx = logger.WithContext(ctx, l)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			conn, err := pubsub.NewConn(cfg.PubSub, "cli.events.pub")
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
			l.Info("load completed")

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
					Type:         args[2],
					Id:           utils.NewId("e"),
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

func NewEventsSub() *cobra.Command {
	command := &cobra.Command{
		Use:               "sub",
		PersistentPreRunE: useChain(),
		Args:              cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			queue := args[0]
			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "events.subscriber").
				With("queue", queue)
			ctx = logger.WithContext(ctx, l)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			conn, err := pubsub.NewConn(cfg.PubSub, "cli.events.sub")
			if err != nil {
				panic(err)
			}
			defer conn.Close()
			ctx = pubsub.WithContext(ctx, conn)
			l.Debugw("load pubsub", "uri", cfg.PubSub.Uri, "stream_name", cfg.PubSub.StreamName, "stream_region", cfg.PubSub.StreamRegion)

			sub, err := app.NewSub(ctx)
			if err != nil {
				log.Fatal(err)
			}
			l.Info("load completed")

			cleanup, err := sub(event.TOPIC_EVENT_PUT, queue, func(msg *pubsub.Message) error {
				var e event.Event
				if err = msgpack.Unmarshal(msg.Data, &e); err != nil {
					l.Errorw(err.Error(), "workspace", msg.Workspace, "app", msg.App, "id", msg.Id)
					return nil
				}

				l.Infow(
					"got event",
					"bucket", e.Bucket,
					"workspace", e.Workspace,
					"app", e.App,
					"type", e.Type,
					"id", e.Id,
					"creation_time", time.UnixMicro(e.CreationTime).Format(time.RFC3339),
				)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
			defer cleanup()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit
		},
	}

	command.Flags().StringArrayP("props", "p", []string{}, "message body properties")

	return command
}
