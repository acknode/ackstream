package cmd

import "github.com/spf13/cobra"

func NewGet() *cobra.Command {
	command := &cobra.Command{
		Use:               "get",
		PersistentPreRunE: Chain(),
	}

	command.AddCommand(NewGetSystem())
	return command
}
