package storage

import (
	"context"
	"time"
)

func Ping(storage *Storage) bool {
	query := storage.Session.Query("SELECT uuid() FROM system.local;")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var id string
	if err := query.WithContext(ctx).Scan(&id); err != nil {
		return false
	}
	return id != ""
}
