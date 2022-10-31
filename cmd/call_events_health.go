package cmd

import (
	"context"
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xrpc"
	"github.com/acknode/ackstream/services/events"
	"github.com/acknode/ackstream/services/events/protos"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

func NewCallEventsHealth() *cobra.Command {
	command := &cobra.Command{
		Use:               "health",
		Short:             "get healthy status of event services",
		Example:           "ackstream call events health",
		PersistentPreRunE: UseChain(),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			cfg := configs.FromContext(ctx)
			logger := xlogger.FromContext(ctx).
				With("cli.fn", "call.events.health").
				With("events.client_remote_address", cfg.XRPC.ClientRemoteAddress)

			conn, err := xrpc.NewClient(ctx, []grpc.DialOption{})
			if err != nil {
				logger.Fatal(err)
			}
			defer func() {
				if err := conn.Close(); err != nil {
					logger.Fatal(err)
				}
			}()

			client, err := events.NewClient(ctx, conn)
			if err != nil {
				logger.Fatal(err)
			}

			meta := metadata.New(map[string]string{
				"content-type":         "application/grpc",
				"acknode-service-name": "ackstream-events",
			})
			logger.Infow("sending", "headers", meta)
			ctx = metadata.NewOutgoingContext(ctx, meta)

			req := &protos.HealthReq{}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			var headers metadata.MD
			res, err := client.Health(ctx, req, grpc.Header(&headers))
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infow("received", "response", res, "headers", headers)
		},
	}

	return command
}
