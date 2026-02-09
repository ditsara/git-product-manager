package cache

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ditsara/git-product-manager/internal/migrations"
)

func RunMigrations(dbPath string) error {
	// Create migration source from embedded filesystem
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migrate instance with embedded migrations
	m, err := migrate.NewWithSourceInstance("iofs", source, "sqlite3://"+dbPath)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Apply all pending migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// openDB opens a database connection (helper for tests)
func openDB(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite3", dbPath)
}
