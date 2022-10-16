package utils

import "os"

func IsDebug(envkey string) bool {
	return os.Getenv(envkey) == "dev"
}
