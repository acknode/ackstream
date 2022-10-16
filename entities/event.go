package entities

import (
	"strings"

	"github.com/acknode/ackstream/utils"
)

type Event struct {
	// partition keys, we often use the timestamp with format YYMMDD
	Bucket    string `json:"bucket"`
	Workspace string `json:"workspace"`
	App       string `json:"app"`
	Type      string `json:"type"`

	// clustering keys
	// chronologically sortable id - ksuid - 1sec resolution
	Id string `json:"id"`

	// properties
	Timestamps int64  `json:"timestamps"`
	Data       string `json:"data"`
}

func (event *Event) SetPartitionKeys(e *Event) bool {
	// only set new partition keys if they are not set yet
	ok := event.Bucket == "" && event.Workspace == "" && event.App == "" && event.Type == ""
	if !ok {
		return false
	}

	event.Bucket = e.Bucket
	event.Workspace = e.Workspace
	event.App = e.App
	event.Type = e.Type
	return true
}

func (event *Event) WithId() bool {
	// only set new id it the id didn't set yet
	if event.Id != "" {
		return false
	}

	event.Id = utils.NewId("event")
	return true
}

func (event *Event) Key() string {
	keys := []string{
		event.Bucket,
		event.Workspace,
		event.App,
		event.Type,
		event.Id,
	}
	return strings.Join(keys, "/")
}
