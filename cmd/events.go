package cmd

import (
	"os"
	"time"

	"github.com/acknode/ackstream/entities"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		PersistentPreRunE: Chain(),
	}

	command.AddCommand(NewEventsPub())
	command.AddCommand(NewEventsSub())
	command.AddCommand(NewEventsGet())

	return command
}

func draw(e *entities.Event, nowrapping bool) {
	t := table.NewWriter()
	if !nowrapping {
		t.SetAllowedRowLength(150)
	}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Key", "Value"})

	t.AppendRow([]interface{}{"bucket", e.Bucket})
	t.AppendRow([]interface{}{"workspace", e.Workspace})
	t.AppendRow([]interface{}{"app", e.App})
	t.AppendRow([]interface{}{"type", e.Type})
	t.AppendRow([]interface{}{"id", e.Id})
	t.AppendRow([]interface{}{"creation_time", time.UnixMicro(e.CreationTime).Format(time.RFC3339)})
	t.AppendSeparator()
	t.AppendRow([]interface{}{"data", e.Data})
	t.Render()
}
