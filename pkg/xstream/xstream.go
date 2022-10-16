package xstream

import (
	"github.com/acknode/ackstream/entities"
)

type SubscribeFn func(event *entities.Event) error

type Sub func(sample *entities.Event, queue string, fn SubscribeFn) error

type Pub func(event *entities.Event) (string, error)
