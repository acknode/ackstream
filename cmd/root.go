package cmd

import (
	"context"

	"github.com/acknode/ackstream/internal/configs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ctxkey string

const CTXKEY_CMD_CONFIGS ctxkey = "ackstream.cmd.configs"

func New() *cobra.Command {
	command := &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cmd.SetContext(context.WithValue(ctx, CTXKEY_CMD_CONFIGS, withConfigs(cmd.Flags())))
		},
	}

	command.PersistentFlags().StringArrayP("configs-dirs", "c", []string{".", "./secrets"}, "path/to/config/file")
	command.PersistentFlags().StringArrayP("set", "s", []string{}, "override value in config file")

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
