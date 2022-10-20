package cmd

import (
	"github.com/spf13/cobra"
)

func NewCallEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		Short:             "call remote APIs for Events",
		Example:           "ackstream call events health",
		ValidArgs:         []string{"health"},
		PersistentPreRunE: Chain(),
	}

	command.AddCommand(NewCallEventsHealth())

	return command
}
