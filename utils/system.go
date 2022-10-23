package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
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

func WithHealthCheck(keyPath string) error {
	return os.WriteFile(keyPath, []byte(time.Now().Format(time.RFC3339)), 0644)
}

func GetRootDir() (*string, error) {
	exec, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(exec)
	cwd := filepath.Join(execDir, "../")
	return &cwd, nil
}
