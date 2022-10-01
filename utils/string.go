package utils

import (
	"fmt"

	"github.com/segmentio/ksuid"
)

func NewId(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, ksuid.New().String())
}
