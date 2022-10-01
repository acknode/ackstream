package cmd

import (
	"context"

	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/zlogger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ctxkey string

const CTXKEY_CONFIGS ctxkey = "ackstream.cmd.configs"
const CTXKEY_LOGGER ctxkey = "ackstream.cmd.logger"

func New() *cobra.Command {
	command := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg := withConfigs(cmd.Flags())
			ctx = context.WithValue(ctx, CTXKEY_CONFIGS, cfg)

			logger := zlogger.New(cfg.Debug)
			ctx = context.WithValue(ctx, CTXKEY_LOGGER, logger)

			cmd.SetContext(ctx)
			return nil
		},
	}

	command.PersistentFlags().StringArrayP("configs-dirs", "c", []string{".", "./secrets"}, "path/to/config/file")
	command.PersistentFlags().StringArrayP("set", "s", []string{}, "override value in config file")

	command.AddCommand(NewStart())
	command.AddCommand(NewEvents())

	return command
}

func withConfigs(flags *pflag.FlagSet) *configs.Configs {
	sets, err := flags.GetStringArray("set")
	if err != nil {
		panic(err)
	}
	cfgdirs, err := flags.GetStringArray("configs-dirs")
	if err != nil {
		panic(err)
	}
	provider, err := configs.NewProvider(cfgdirs...)
	if err != nil {
		panic(err)
	}

	cfg, err := configs.New(provider, sets)
	if err != nil {
		panic(err)
	}

	return cfg
}

func clichain() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		parent := cmd.Parent()
		err := parent.PersistentPreRunE(parent, args)

		cmd.SetContext(parent.Context())
		return err
	}
}
