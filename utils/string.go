package utils

import (
	"fmt"
	"github.com/Machiel/slugify"
	"time"

	"github.com/segmentio/ksuid"
)

func NewId(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, ksuid.New().String())
}

func NewBucket(template string) (string, int64) {
	now := time.Now().UTC()
	return now.Format(template), now.UnixMilli()
}

func SnakeCase(value string) string {
	slug := slugify.New(slugify.Configuration{ReplaceCharacter: '_'})
	return slug.Slugify(value)
}
