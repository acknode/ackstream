package cmd

import (
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func New() *cobra.Command {
	command := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := WithConfigs(cmd.Flags())
			if err != nil {
				return err
			}
			ctx = configs.WithContext(ctx, cfg)

			logger := xlogger.New(cfg.Debug)
			ctx = xlogger.WithContext(ctx, logger)

			cmd.SetContext(ctx)
			return nil
		},
		ValidArgs: []string{"get", "pub", "sub"},
	}

	command.PersistentFlags().StringArrayP("configs-dirs", "c", []string{".", "./secrets"}, "path/to/config/file")
	command.PersistentFlags().StringArrayP("set", "s", []string{}, "override values in config file")

	command.AddCommand(NewGet())
	command.AddCommand(NewPub())
	return command
}

func WithConfigs(flags *pflag.FlagSet) (*configs.Configs, error) {
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

func Chain() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		parent := cmd.Parent()
		err := parent.PersistentPreRunE(parent, args)

		cmd.SetContext(parent.Context())
		return err
	}
}
