package cmd

import (
	"github.com/spf13/cobra"
)

func NewCall() *cobra.Command {
	command := &cobra.Command{
		Use:               "call",
		Short:             "call remote APIs",
		Example:           "ackstream call events",
		ValidArgs:         []string{"events"},
		PersistentPreRunE: UseChain(),
	}

	command.AddCommand(NewCallEvents())

	return command
}
