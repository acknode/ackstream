package cmd

import "github.com/spf13/cobra"

func NewGet() *cobra.Command {
	command := &cobra.Command{
		Use:               "get",
		Short:             "display one or many resources",
		Example:           "ackstream get system",
		ValidArgs:         []string{"system"},
		PersistentPreRunE: UseChain(),
	}

	command.AddCommand(NewGetSystem())
	command.AddCommand(NewGetEvents())

	return command
}
