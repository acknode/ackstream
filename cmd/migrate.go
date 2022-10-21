package cmd

import (
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/spf13/cobra"
)

func NewMigrate() *cobra.Command {
	command := &cobra.Command{
		Use:               "migrate directory",
		Short:             "run migration for our storage",
		Example:           "ackstream migrate ./migrate",
		Args:              cobra.MinimumNArgs(1),
		PersistentPreRunE: UseChain(),
		Run: func(cmd *cobra.Command, args []string) {
			logger := xlogger.FromContext(cmd.Context()).With("cli.fn", "migrate")

			if err := xstorage.Migrate(cmd.Context(), args); err != nil {
				logger.Error(err)
			}
		},
	}

	return command
}
