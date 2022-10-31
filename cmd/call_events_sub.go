package cmd

import (
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xrpc"
	"github.com/acknode/ackstream/services/events"
	"github.com/acknode/ackstream/services/events/protos"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
)

func NewCallEventsSub() *cobra.Command {
	command := &cobra.Command{
		Use:               "sub",
		Short:             "subscribe event streaming from our gRPC service",
		Example:           "ackstream call events sub -w ws_default -a app_demo -t cli.grpc.trigger",
		PersistentPreRunE: UseChain(),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			cfg := configs.FromContext(ctx)
			logger := xlogger.FromContext(ctx).
				With("cli.fn", "call.events.sub").
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

			event, err := parseEvent(cmd.Flags())
			if err != nil {
				logger.Fatal(err)
			}

			meta := metadata.New(map[string]string{
				"content-type":         "application/grpc",
				"acknode-service-name": "ackstream-events",
			})
			logger.Infow("sending", "event_key", event.Key(), "headers", meta)
			ctx = metadata.NewOutgoingContext(ctx, meta)

			req := &protos.SubReq{
				Workspace: event.Workspace,
				App:       event.App,
				Type:      event.Type,
			}

			var headers metadata.MD
			stream, err := client.Sub(ctx, req, grpc.Header(&headers))
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infow("receiving", "headers", headers)

			for {
				event, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					logger.Error(err)
				}

				logger.Infow("received", "event", event)
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
