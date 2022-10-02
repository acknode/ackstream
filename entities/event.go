package entities

import (
	"strings"

	"github.com/acknode/ackstream/utils"
)

type Event struct {
	// partition keys
	// often be the timestamp with format YYMMDD
	Bucket    string `json:"bucket"`
	Workspace string `json:"workspace"`
	App       string `json:"app"`
	Type      string `json:"type"`

	// clustering keys
	// chronologically sortable id - ksuid - 1-second resolution
	Id string `json:"id"`

	// properties
	CreationTime int64  `json:"creation_time"`
	Data         string `json:"data"`
}

func (event *Event) WithId() bool {
	// only set new id it the id didn't set yet
	if event.Id == "" {
		event.Id = utils.NewId("event")
		return false
	}

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
