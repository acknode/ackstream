package cmd

import (
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewCallEventsHealth() *cobra.Command {
	command := &cobra.Command{
		Use:               "health",
		Short:             "get healthy status of event services",
		Example:           "ackstream call events health",
		PersistentPreRunE: Chain(),
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "call.events.health")

			cfg, err := parseEventsCfg(cmd.Flags())
			if err != nil {
				logger.Fatal(err)
			}

			transportOpts := grpc.WithTransportCredentials(insecure.NewCredentials())
			client, err := grpc.Dial(cfg.GRPCListenAddress, transportOpts)
			if err != nil {
				logger.Fatal(err)
			}

			service := proto.NewEventsClient(client)
			req := &proto.HealthReq{}

			res, err := service.Health(cmd.Context(), req)
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infow("received", "response", res)

			if err := client.Close(); err != nil {
				logger.Fatal(err)
			}
		},
	}

	return command
}
