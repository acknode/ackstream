package cmd

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/acknode/ackstream/entities"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewEvents() *cobra.Command {
	command := &cobra.Command{
		Use:               "events",
		PersistentPreRunE: Chain(),
	}

	command.AddCommand(NewEventsPub())
	command.AddCommand(NewEventsSub())
	command.AddCommand(NewEventsGet())

	command.PersistentFlags().StringP("workspace", "w", "", "the sample workspace you want to listen events")
	command.PersistentFlags().StringP("app", "a", "", "the sample application you want to listen events")
	command.PersistentFlags().StringP("type", "t", "", "the sample type you want to listen events")

	return command
}

func draw(e *entities.Event, nowrapping bool) {
	t := table.NewWriter()
	if !nowrapping {
		t.SetAllowedRowLength(80)
	}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Key", "Value"})

	t.AppendRow([]interface{}{"bucket", e.Bucket})
	t.AppendRow([]interface{}{"workspace", e.Workspace})
	t.AppendRow([]interface{}{"app", e.App})
	t.AppendRow([]interface{}{"type", e.Type})
	t.AppendRow([]interface{}{"id", e.Id})
	t.AppendRow([]interface{}{"creation_time", time.UnixMilli(e.CreationTime).Format(time.RFC3339)})
	t.AppendSeparator()
	t.AppendRow([]interface{}{"data", e.Data})
	t.AppendRow([]interface{}{"length", humanize.Bytes(uint64(len([]byte(e.Data))))})
	t.Render()
}

func getSampleEvent(flags *pflag.FlagSet, required bool) *entities.Event {
	var event entities.Event

	if ws, err := flags.GetString("workspace"); err == nil && ws != "" {
		event.Workspace = ws
	}
	if app, err := flags.GetString("app"); err == nil && app != "" {
		event.App = app
	}
	if etype, err := flags.GetString("type"); err == nil && etype != "" {
		event.Type = etype
	}

	ok := event.Workspace != "" && event.App != "" && event.Type != ""
	if ok {
		return &event
	}

	if required {
		log.Fatal(errors.New("neither workspace nor app nor type were not set"))
		return nil
	}

	return &event
}
