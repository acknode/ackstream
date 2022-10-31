package cmd

import (
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xrpc"
	"github.com/acknode/ackstream/services/events"
	"github.com/acknode/ackstream/services/events/protos"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
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
				With("cli.fn", "call.events.pub").
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

			req := &protos.HealthReq{}
			res, err := client.Health(cmd.Context(), req)
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infow("received", "response", res)
		},
	}

	return command
}
