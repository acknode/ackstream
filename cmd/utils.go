package cmd

import (
	"encoding/json"
	"strings"

	"github.com/acknode/ackstream/entities"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/pflag"
	"os"
	"time"
)

func parseEventSample(flags *pflag.FlagSet) *entities.Event {
	event := &entities.Event{}

	if bucket, err := flags.GetString("bucket"); err == nil {
		event.Bucket = bucket
	}
	if ws, err := flags.GetString("workspace"); err == nil {
		event.Workspace = ws
	}
	if app, err := flags.GetString("app"); err == nil {
		event.App = app
	}
	if etype, err := flags.GetString("type"); err == nil {
		event.Type = etype
	}
	if id, err := flags.GetString("id"); err == nil {
		event.Id = id
	}

	return event
}

func parseEvent(flags *pflag.FlagSet) (*entities.Event, error) {
	event := parseEventSample(flags)

	data := map[string]interface{}{}
	props, err := flags.GetStringArray("props")
	if err != nil {
		return nil, err
	}
	for _, prop := range props {
		kv := strings.Split(prop, "=")
		data[kv[0]] = kv[1]
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	event.Data = string(bytes)

	return event, nil
}

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
