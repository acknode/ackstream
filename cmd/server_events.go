package cmd

import (
	"context"
	"github.com/acknode/ackstream/app"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events"
	"github.com/acknode/ackstream/services/events/configs"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewServeEvents() *cobra.Command {
	command := &cobra.Command{
		Use:     "events",
		Short:   "serve events service",
		Example: "ackstream serve events",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chain := UseChain()
			if err := chain(cmd, args); err != nil {
				return err
			}

			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "serve.events")
			cfg, err := parseEventsCfg(cmd.Flags())
			if err != nil {
				logger.Fatal(err)
			}
			ctx := configs.WithContext(cmd.Context(), cfg)
			cmd.SetContext(ctx)
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := configs.FromContext(cmd.Context())
			logger := xlogger.FromContext(cmd.Context()).
				With("cli.fn", "serve.events").
				With("events.grpc_listen_address", cfg.GRPCListenAddress)

			ctx, err := app.Connect(cmd.Context())
			if err != nil {
				logger.Fatal(err)
			}
			defer func() {
				if _, err := app.Disconnect(ctx); err != nil {
					logger.Error(err)
				}
			}()

			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			server, err := events.NewServer(ctx)
			if err != nil {
				logger.Fatal(err)
			}

			go func() {
				listener, err := net.Listen("tcp", cfg.GRPCListenAddress)
				if err != nil {
					logger.Fatal(err)
				}
				if err := server.Serve(listener); err != nil {
					logger.Fatal(err)
				}
			}()
			logger.Infow("started gRPC", "endpoint", cfg.GRPCListenAddress)

			if err := utils.WithHealthCheck(events.HEALTHCHECK_FILEPATH); err != nil {
				logger.Fatal(err)
			}
			logger.Infow("set healthcheck", "filepath", events.HEALTHCHECK_FILEPATH)

			// Listen for the interrupt signal.
			<-ctx.Done()
			stop()
			logger.Info("shutting down gracefully, press Ctrl+C again to force")
			// The context is used to inform the server it has 5 seconds to finish
			// the request it is currently handling
			ctx, cancel := context.WithTimeout(ctx, 11*time.Second)
			defer cancel()

			go func() {
				server.GracefulStop()

				if _, err = app.Disconnect(ctx); err != nil {
					logger.Error(err)
				}
				<-ctx.Done()
			}()
		},
	}

	return command
}

func parseEventsCfg(flags *pflag.FlagSet) (*configs.Configs, error) {
	sets, err := flags.GetStringArray("set")
	if err != nil {
		return nil, err
	}
	cfgdirs, err := flags.GetStringArray("configs-dirs")
	if err != nil {
		return nil, err
	}

	provider, err := configs.NewProvider(cfgdirs...)
	if err != nil {
		return nil, err
	}

	cfg, err := configs.New(provider, sets)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
