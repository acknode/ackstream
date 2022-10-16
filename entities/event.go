package entities

import (
	"github.com/go-playground/validator/v10"
	"strings"

	"github.com/acknode/ackstream/utils"
)

type Event struct {
	// partition keys, we often use the timestamp with format YYMMDD
	Bucket    string `json:"bucket" validate:"required"`
	Workspace string `json:"workspace" validate:"required"`
	App       string `json:"app" validate:"required"`
	Type      string `json:"type" validate:"required"`

	// clustering keys
	// chronologically sortable id - ksuid - 1sec resolution
	Id string `json:"id" validate:"required"`

	// properties
	Timestamps int64                  `json:"timestamps" validate:"required,gt=0"`
	Data       map[string]interface{} `json:"data" validate:"required"`
}

func (event *Event) SetPartitionKeys(e *Event) bool {
	// only set new partition keys if they are not set yet
	notset := event.Bucket == "" && event.Workspace == "" && event.App == "" && event.Type == ""
	if !notset {
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

func (event *Event) Valid() bool {
	validate := validator.New()
	return validate.Struct(event) == nil
}
