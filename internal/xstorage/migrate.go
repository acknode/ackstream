package xstorage

import (
	"fmt"

	"github.com/gocql/gocql"
)

func Migrate(cfg *Configs) error {
	cluster := gocql.NewCluster(cfg.Hosts...)
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	keyspaceql := fmt.Sprintf(
		`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1};`,
		cfg.Keyspace,
	)
	if err := session.Query(keyspaceql).Exec(); err != nil {
		return err
	}

	tableql := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s.%s (
			bucket TEXT, workspace TEXT, app TEXT, type TEXT, id TEXT, 
			data BLOB, creation_time BIGINT, 
			PRIMARY KEY ((bucket, workspace, app, type), id)) 
			WITH CLUSTERING ORDER BY (id DESC);`,
		cfg.Keyspace, cfg.Table)
	if err := session.Query(tableql).Exec(); err != nil {
		return err
	}

	return nil
}
