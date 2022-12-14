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
	Timestamps int64  `json:"timestamps"`
	Data       string `json:"data"`
}

func (event *Event) WithId() error {
	// only set data if it wasn't set yet
	if event.Id != "" {
		return ErrEventIdWasSet
	}

	event.Id = utils.NewId("event")
	return nil
}

func (event *Event) WithBucket(template string) error {
	// only set data if it wasn't set yet
	if event.Timestamps > 0 {
		return ErrEventTsWasSet
	}
	if event.Bucket != "" {
		return ErrEventBucketWasSet
	}
	event.Bucket, event.Timestamps = utils.NewBucket(template)
	return nil
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
