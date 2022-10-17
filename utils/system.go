package utils

import (
	"io/fs"
	"os"
	"path/filepath"
)

func IsDebug(key string) bool {
	return os.Getenv(key) == "dev"
}

func ScanFiles(root, ext string) ([]string, error) {
	var filepaths []string
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			filepaths = append(filepaths, s)
		}
		return nil
	})
	return filepaths, err
}
