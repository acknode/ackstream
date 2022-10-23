package entities

import "errors"

var (
	ErrEventIdWasSet     = errors.New("event id has set already")
	ErrEventTsWasSet     = errors.New("event timestamps has set already")
	ErrEventBucketWasSet = errors.New("event bucket has set already")
)
