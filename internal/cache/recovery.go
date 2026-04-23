package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// isRecoverableError reports whether err is a known database error that can be
// resolved by deleting and rebuilding the cache database.
func isRecoverableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table") ||
		strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "file is not a database") ||
		strings.Contains(msg, "unable to open database file")
}

// RecoverFromError checks whether err is a known recoverable database error.
// If it is, it deletes the cache database, re-runs migrations to create a fresh
// schema, and syncs from the filesystem. Returns nil only if recovery succeeded.
// Returns the original err unchanged for non-recoverable errors, or a recovery
// error if the rebuild itself fails.
func RecoverFromError(err error, pmPath string) error {
	if err == nil {
		return nil
	}
	if !isRecoverableError(err) {
		return err
	}

	fmt.Fprintln(os.Stderr, "⚠  Cache corrupted, rebuilding...")

	dbPath := filepath.Join(pmPath, ".cache.db")
	if removeErr := os.Remove(dbPath); removeErr != nil && !os.IsNotExist(removeErr) {
		return fmt.Errorf("failed to remove corrupted cache: %w", removeErr)
	}

	if migrateErr := RunMigrations(dbPath); migrateErr != nil {
		return fmt.Errorf("failed to initialize cache after recovery: %w", migrateErr)
	}

	if syncErr := SyncCache(pmPath); syncErr != nil {
		return fmt.Errorf("failed to sync cache after recovery: %w", syncErr)
	}

	fmt.Fprintln(os.Stderr, "✓ Cache rebuilt successfully")
	return nil
}
