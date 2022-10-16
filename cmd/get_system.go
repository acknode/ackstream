package cmd

import (
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/spf13/cobra"
)

func NewGetSystem() *cobra.Command {
	command := &cobra.Command{
		Use:               "system",
		PersistentPreRunE: Chain(),
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context())

			cfg := configs.FromContext(cmd.Context())
			logger.Infow("common configs", "debug", cfg.Debug, "version", cfg.Version)
		},
	}

	return command
}
