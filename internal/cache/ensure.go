package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureCacheReady ensures the SQLite cache database exists and has the latest schema.
// This function is idempotent and safe to call before any cache operation.
//
// It will:
// - Create the database and run migrations if .cache.db doesn't exist
// - Run any pending migrations if the schema is outdated
// - Return nil if the database is already up-to-date (fast path)
func EnsureCacheReady(pmPath string) error {
	dbPath := filepath.Join(pmPath, ".cache.db")
	
	// Find migrations directory
	migrationPath := FindMigrationPath()
	if migrationPath == "" {
		return fmt.Errorf("migrations directory not found")
	}
	
	// Run migrations (idempotent - safe to call every time)
	// If database doesn't exist, it will be created
	// If schema is current, this is a fast no-op
	if err := RunMigrations(dbPath, migrationPath); err != nil {
		return fmt.Errorf("failed to ensure cache schema: %w", err)
	}
	
	return nil
}

// FindMigrationPath locates the migration directory
// Tries: 1) relative to current dir, 2) relative to executable, 3) walk up tree
func FindMigrationPath() string {
	// Try relative to current directory (for development)
	if _, err := os.Stat("migrations"); err == nil {
		absPath, _ := filepath.Abs("migrations")
		return absPath
	}

	// Try relative to executable (for installed binary)
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		migrationPath := filepath.Join(execDir, "migrations")
		if _, err := os.Stat(migrationPath); err == nil {
			return migrationPath
		}

		// Try one level up (for bin/pm structure)
		migrationPath = filepath.Join(execDir, "..", "migrations")
		if _, err := os.Stat(migrationPath); err == nil {
			absPath, _ := filepath.Abs(migrationPath)
			return absPath
		}
	}

	// Walk up the directory tree (for tests running in subdirectories)
	for i := 1; i <= 5; i++ {
		path := filepath.Join(strings.Repeat("../", i), "migrations")
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return ""
}
