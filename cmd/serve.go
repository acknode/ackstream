package cmd

import "github.com/spf13/cobra"

func NewServe() *cobra.Command {
	command := &cobra.Command{
		Use:               "serve",
		Short:             "serve a service",
		Example:           "ackstream serve events",
		ValidArgs:         []string{"events"},
		PersistentPreRunE: UseChain(),
	}

	command.AddCommand(NewServeEvents())

	return command
}
