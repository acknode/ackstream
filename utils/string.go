package utils

import (
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
)

func NewId(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, ksuid.New().String())
}

func NewBucket(t time.Time) string {
	return t.Format("20060102")
}
