package cmd

import (
	"github.com/acknode/ackstream/entities"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"time"
)

func printEvent(event *entities.Event) {
	t := table.NewWriter()
	t.SetAllowedRowLength(80)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Key", "Value"})

	t.AppendRow([]interface{}{"bucket", event.Bucket})
	t.AppendRow([]interface{}{"workspace", event.Workspace})
	t.AppendRow([]interface{}{"app", event.App})
	t.AppendRow([]interface{}{"type", event.Type})
	t.AppendRow([]interface{}{"id", event.Id})
	t.AppendRow([]interface{}{"timestamps", time.UnixMilli(event.Timestamps).Format(time.RFC3339)})
	t.AppendSeparator()
	t.AppendRow([]interface{}{"data", event.Data})
	t.Render()
}
