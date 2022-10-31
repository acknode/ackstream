package cmd

import (
	"github.com/spf13/cobra"
)

func NewCallEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		Short:             "call remote APIs for Events",
		Example:           "ackstream call events health",
		ValidArgs:         []string{"health", "pub"},
		PersistentPreRunE: UseChain(),
	}

	command.AddCommand(NewCallEventsHealth())
	command.AddCommand(NewCallEventsPub())
	command.AddCommand(NewCallEventsSub())

	return command
}
