package cmd

import (
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events"
	"github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/protos"
	"github.com/spf13/cobra"
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
				With("events.grpc_listen_address", cfg.GRPCListenAddress)

			conn, client, err := events.NewClient(ctx)
			if err != nil {
				logger.Fatal(err)
			}

			req := &protos.HealthReq{}
			res, err := client.Health(cmd.Context(), req)
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infow("received", "response", res)

			if err := conn.Close(); err != nil {
				logger.Fatal(err)
			}
		},
	}

	return command
}
