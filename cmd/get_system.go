package cmd

import (
	"github.com/acknode/ackstream/internal/configs"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"os"
)

func NewGetSystem() *cobra.Command {
	command := &cobra.Command{
		Use:               "system",
		Short:             "show system information & configuration",
		Example:           "ackstream get system",
		PersistentPreRunE: Chain(),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := configs.FromContext(cmd.Context())

			t := table.NewWriter()
			t.SetAllowedRowLength(80)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Key", "Value"})

			t.AppendRow([]interface{}{"debug", cfg.Debug})
			t.AppendRow([]interface{}{"version", cfg.Version})
			t.Render()
		},
	}

	return command
}
