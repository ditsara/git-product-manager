package cache

import (
	"fmt"
	"path/filepath"
)

// EnsureCacheReady ensures the SQLite cache database exists and has the latest schema.
// This function is idempotent and safe to call before any cache operation.
//
// It will:
// - Create the database and run migrations if .cache.db doesn't exist
// - Run any pending migrations if the schema is outdated
// - Return nil if the database is already up-to-date (fast path)
// - Automatically recover from known corruption errors (at most once per call)
func EnsureCacheReady(pmPath string) error {
	dbPath := filepath.Join(pmPath, ".cache.db")

	// Run migrations (idempotent - safe to call every time).
	// If database doesn't exist, it will be created.
	// If schema is current, this is a fast no-op.
	if err := RunMigrations(dbPath); err != nil {
		// Attempt auto-recovery for known corruption/lock errors.
		// RecoverFromError deletes the DB, re-runs migrations, and syncs
		// from the filesystem. Returns nil only if recovery succeeded.
		if recoverErr := RecoverFromError(err, pmPath); recoverErr != nil {
			return fmt.Errorf("failed to ensure cache schema: %w", recoverErr)
		}
		// Recovery succeeded — the DB is fresh and already synced.
	}

	return nil
}

