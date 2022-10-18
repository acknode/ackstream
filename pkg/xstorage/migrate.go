package xstorage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/acknode/ackstream/pkg/xlogger"
	"github.com/acknode/ackstream/utils"
	"github.com/gocql/gocql"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"
)

func Migrate(ctx context.Context, dirs []string) error {
	logger := xlogger.FromContext(ctx).
		With("pkg", "xstorage").
		With("fn", "migrate")
	pattern := ".cql"

	if len(dirs) == 0 {
		return nil
	}

	var filepaths []string
	for _, p := range dirs {
		paths, err := utils.ScanFiles(p, pattern)
		if err != nil {
			logger.Error(err)
			continue
		}

		filepaths = append(filepaths, paths...)
	}

	if len(filepaths) == 0 {
		return fmt.Errorf("no migration files found with glob %s", pattern)
	}
	sort.Strings(filepaths)

	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return err
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	conn, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, filepath := range filepaths {
		if err := migrate(conn, cfg, filepath); err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func migrate(conn *gocql.Session, cfg *Configs, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	i := 0
	r := bytes.NewBuffer(b)
	for {
		stmt, err := r.ReadString(';')
		if err == io.EOF {
			// handle missing semicolon after last statement
			if strings.TrimSpace(stmt) != "" {
				err = nil
			} else {
				break
			}
		}
		if err != nil {
			return err
		}
		i++

		tmpl, err := template.New(fmt.Sprintf("%s.migrate.%d", filepath, i)).Parse(strings.TrimSpace(stmt))
		if err != nil {
			return err
		}

		var cql bytes.Buffer
		if err := tmpl.Execute(&cql, cfg); err != nil {
			return err
		}

		if err := conn.Query(cql.String()).Exec(); err != nil {
			return err
		}
	}
	return file.Close()
}
