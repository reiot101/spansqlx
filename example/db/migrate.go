package db

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/spanner"

	// migrate driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
)

type MigrateOperation int8

const (
	UP MigrateOperation = iota
	DOWN
)

func Migrate(uri string, path string, op MigrateOperation) error {
	s := new(spanner.Spanner)
	d, err := s.Open(uri + "?x-clean-statements=true")
	if err != nil {
		return err
	}

	// path options
	// https://github.com/golang-migrate/migrate/tree/master/source/file
	// https://github.com/golang-migrate/migrate/tree/master/source/github
	m, err := migrate.NewWithDatabaseInstance(path, uri, d)
	if err != nil {
		return err
	}

	// migrate up/down operations
	switch op {
	case UP:
		if err := m.Up(); err != nil {
			// Already up operations then skip
			if errors.Is(err, migrate.ErrNoChange) {
				return nil
			}
			return err
		}
	case DOWN:
		if err := m.Down(); err != nil {
			return err
		}
	}

	log.Printf("%s migrated.\n", path)

	return nil
}
