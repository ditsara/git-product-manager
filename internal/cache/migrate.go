package cache

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbPath string, migrationPath string) error {
	// The migration path needs to be prefixed with "file://"
	m, err := migrate.New(
		"file://"+migrationPath,
		"sqlite3://"+dbPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// openDB opens a database connection (helper for tests)
func openDB(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite3", dbPath)
}
