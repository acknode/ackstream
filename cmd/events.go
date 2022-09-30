package cmd

import (
	"github.com/spf13/cobra"
)

func NewEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		PersistentPreRunE: clichain(),
	}

	command.AddCommand(NewEventsPub())
	command.AddCommand(NewEventsSub())
	command.AddCommand(NewEventsGet())

	return command
}
