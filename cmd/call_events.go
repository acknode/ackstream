package cmd

import (
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/services/events/configs"
	"github.com/spf13/cobra"
)

func NewCallEvents() *cobra.Command {
	command := &cobra.Command{
		Use:       "events",
		Short:     "call remote APIs for Events",
		Example:   "ackstream call events health",
		ValidArgs: []string{"health", "pub"},
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
	}

	command.AddCommand(NewCallEventsHealth())
	command.AddCommand(NewCallEventsPub())

	return command
}
