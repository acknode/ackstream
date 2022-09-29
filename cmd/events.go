package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/event"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/logger"
	"github.com/acknode/ackstream/internal/xstorage"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/cobra"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

func NewEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		PersistentPreRunE: clichain(),
	}

	command.AddCommand(NewEventsPub())
	command.AddCommand(NewEventsSub())
	command.AddCommand(NewEventsGet())

	return command
}

func NewEventsPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub",
		PersistentPreRunE: clichain(),
		Args:              cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "cli.events.publisher")
			ctx = logger.WithContext(ctx, l)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			pub, cb := app.UsePub(ctx)
			defer cb()
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
			data, err := msgpack.Marshal(payload)
			if err != nil {
				log.Fatal(err)
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

			l.Infow("published", "publish_key", pubkey)
		},
	}

	command.Flags().StringArrayP("props", "p", []string{}, "message body properties")

	return command
}

func NewEventsSub() *cobra.Command {
	command := &cobra.Command{
		Use:               "sub",
		PersistentPreRunE: clichain(),
		Args:              cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			queue := args[0]
			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "cli.events.subscriber").
				With("queue", queue)
			ctx = logger.WithContext(ctx, l)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			cb, err := app.UseSub(ctx, queue, func(e *event.Event) error {
				log.Printf("event: bucket=%s ws=%s app=%s type=%s id=%s", e.Bucket, e.Workspace, e.App, e.Type, e.Id)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
			defer cb()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit
		},
	}

	command.Flags().StringArrayP("props", "p", []string{}, "message body properties")

	return command
}

func NewEventsGet() *cobra.Command {
	command := &cobra.Command{
		Use:               "get",
		PersistentPreRunE: clichain(),
		Args:              cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			l := cmd.Context().
				Value(CTXKEY_LOGGER).(*zap.SugaredLogger).
				With("service", "cli.events.get")
			ctx = logger.WithContext(ctx, l)

			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			ctx = configs.WithContext(ctx, cfg)
			l.Debugw("load configs", "version", cfg.Version, "debug", cfg.Debug)

			session := xstorage.New(ctx, cfg.Storage)
			defer session.Close()
			ctx = xstorage.WithContext(ctx, session)
			l.Debugw("load storage", "hosts", cfg.Storage.Hosts, "keyspace", cfg.Storage.Keyspace, "table", cfg.Storage.Table)

			get := xstorage.UseGet(ctx, cfg.Storage)
			e, err := get(args[0], args[1], args[2], args[3], args[4])
			if err != nil {
				l.Fatal(err)
			}

			var data interface{}
			if err := msgpack.Unmarshal(e.Data, &data); err != nil {
				l.Fatal(err)
			}

			l.Infow("get event",
				"key", e.Key(),
				"data", data,
				"creation_time", time.UnixMicro(e.CreationTime).Format(time.RFC3339),
			)
		},
	}

	return command
}
