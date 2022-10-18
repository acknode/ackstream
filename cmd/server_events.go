package cmd

import (
	"context"
	"fmt"
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
		Use:               "events",
		Short:             "serve events service",
		Example:           "ackstream serve events",
		PersistentPreRunE: Chain(),
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "serve.events")

			ctx, err := app.Connect(cmd.Context())
			if err != nil {
				logger.Fatal(err)
			}
			defer func() {
				if _, err := app.Disconnect(ctx); err != nil {
					logger.Error(err)
				}
			}()

			server, err := events.New(ctx)
			if err != nil {
				logger.Fatal(err)
			}

			cfg, err := parseEventsCfg(cmd.Flags())
			if err != nil {
				logger.Fatal(err)
			}
			address := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
			listener, err := net.Listen("tcp", address)
			if err != nil {
				logger.Fatal(err)
			}

			logger.Infof("listening... %s", address)
			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			go func() {
				if err := server.Serve(listener); err != nil {
					logger.Fatal(err)
				}
				if err := utils.WithHealthCheck("/tmp/ackstream.services.events"); err != nil {
					logger.Fatal(err)
				}
			}()

			// Listen for the interrupt signal.
			<-ctx.Done()
			stop()
			logger.Info("shutting down gracefully, press Ctrl+C again to force")
			// The context is used to inform the server it has 5 seconds to finish
			// the request it is currently handling
			ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
			defer cancel()

			go func() {
				server.GracefulStop()
				_, _ = app.Disconnect(ctx)
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
