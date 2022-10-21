package cmd

import (
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events"
	eventscfg "github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/services/events/protos"
	"github.com/spf13/cobra"
)

func NewCallEventsPub() *cobra.Command {
	command := &cobra.Command{
		Use:               "pub",
		Short:             "publish an event to our gRPC service",
		Example:           "ackstream call events pub -w ws_default -a app_demo -t cli.grpc.trigger -p env=local -p os=macos",
		PersistentPreRunE: UseChain(),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			cfg := eventscfg.FromContext(ctx)
			logger := xlogger.FromContext(ctx).
				With("cli.fn", "call.events.pub").
				With("events.grpc_listen_address", cfg.GRPCListenAddress)

			conn, client, err := events.NewClient(ctx)
			if err != nil {
				logger.Fatal(err)
			}

			event, err := parseEvent(cmd.Flags())
			if err != nil {
				logger.Fatal(err)
			}

			req := &protos.PubReq{
				Workspace: event.Workspace,
				App:       event.App,
				Type:      event.Type,
				Data:      event.Data,
			}
			res, err := client.Pub(ctx, req)
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infow("received", "response", res)

			if err := conn.Close(); err != nil {
				logger.Fatal(err)
			}
		},
	}

	command.Flags().StringP("workspace", "w", "", " --workspace='': specify which workspace you want to publish an event to")
	_ = command.MarkFlagRequired("workspace")

	command.Flags().StringP("app", "a", "", "--app='': specify which app you are working with")
	_ = command.MarkFlagRequired("app")

	command.Flags().StringP("type", "t", "", "--type='': specify which type of event you want to use")
	_ = command.MarkFlagRequired("type")

	command.Flags().StringArrayP("props", "p", []string{}, "--props='env=local': set data properties")

	return command
}
