package events

import "errors"

var (
	ErrEventNoWorkspace = errors.New("event workspace is required")
	ErrEventNoApp       = errors.New("event app is required")
	ErrEventNoType      = errors.New("event type is required")
)
