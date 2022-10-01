package cmd

import (
	"github.com/acknode/ackstream/internal/configs"
	"github.com/acknode/ackstream/internal/xstorage"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewStart() *cobra.Command {
	command := &cobra.Command{
		Use: "start",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chain := Chain()
			if err := chain(cmd, args); err != nil {
				return err
			}

			auto, err := cmd.Flags().GetBool("auto-migrate")
			if err != nil {
				return err
			}

			l := cmd.Context().Value(CTXKEY_LOGGER).(*zap.SugaredLogger)
			cfg := cmd.Context().Value(CTXKEY_CONFIGS).(*configs.Configs)
			if auto && !cfg.Debug {
				l.Warnw("set auto migrate but environment is not development")
			}
			// migrate storage before start
			if auto && cfg.Debug {
				l.Debugw("migrating", "hosts", cfg.XStorage.Hosts, "keyspace", cfg.XStorage.Keyspace, "table", cfg.XStorage.Table)
				return xstorage.Migrate(cfg.XStorage)
			}

			return nil
		},
	}

	command.AddCommand(NewStartDatastore())
	command.PersistentFlags().BoolP("auto-migrate", "", false, "run auto-migration process when start an application")

	return command
}
