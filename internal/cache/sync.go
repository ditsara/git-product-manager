package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
	_ "github.com/mattn/go-sqlite3"
)

// ShouldSync checks if the cache needs to be synchronized with the filesystem
// by comparing the last sync timestamp with the modification times of ticket files
func ShouldSync(pmPath string) (bool, error) {
	dbPath := filepath.Join(pmPath, ".cache.db")
	
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get last sync timestamp
	var lastSyncStr string
	err = db.QueryRow("SELECT value FROM cache_metadata WHERE key = 'last_sync_timestamp'").Scan(&lastSyncStr)
	if err != nil {
		// If metadata doesn't exist, we need to sync
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, fmt.Errorf("failed to get last sync timestamp: %w", err)
	}

	lastSync, err := time.Parse(time.RFC3339, lastSyncStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse last sync timestamp: %w", err)
	}

	// Check if any ticket file is newer than last sync
	ticketsPath := filepath.Join(pmPath, "tickets")
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		// If tickets directory doesn't exist, no need to sync
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read tickets directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// TIMING NOTE: Use truncated comparison to account for filesystem timestamp precision.
		// Most filesystems store mtime with second-level precision, but Go's time.Time uses
		// nanosecond precision. Truncating both times to seconds ensures we don't get false
		// positives from subsecond differences between when we record the sync time and when
		// the filesystem actually stored the file's mtime.
		fileTime := info.ModTime().Truncate(time.Second)
		syncTime := lastSync.Truncate(time.Second)
		
		if fileTime.After(syncTime) {
			return true, nil
		}
	}

	return false, nil
}

// SyncCache synchronizes the SQLite cache with the filesystem
// by scanning all ticket files and updating the database
func SyncCache(pmPath string) error {
	dbPath := filepath.Join(pmPath, ".cache.db")
	
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing tickets
	_, err = tx.Exec("DELETE FROM tickets")
	if err != nil {
		return fmt.Errorf("failed to clear tickets table: %w", err)
	}

	// Scan tickets directory
	ticketsPath := filepath.Join(pmPath, "tickets")
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No tickets yet, just update timestamp
			return updateSyncTimestamp(tx)
		}
		return fmt.Errorf("failed to read tickets directory: %w", err)
	}

	// Prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO tickets (id, title, type, status, priority, assignee, parent, created_at, updated_at, body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Process each ticket file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(ticketsPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		t, err := ticket.Parse(content)
		if err != nil {
			continue // Skip invalid tickets
		}

		// Extract body (everything after the second ---)
		parts := strings.SplitN(string(content), "---", 3)
		body := ""
		if len(parts) == 3 {
			body = strings.TrimSpace(parts[2])
		}

		_, err = stmt.Exec(
			t.ID,
			t.Title,
			t.Type,
			t.Status,
			t.Priority,
			t.Assignee,
			t.Parent,
			t.CreatedAt,
			t.UpdatedAt,
			body,
		)
		if err != nil {
			return fmt.Errorf("failed to insert ticket %s: %w", t.ID, err)
		}
	}

	// Update sync timestamp
	if err := updateSyncTimestamp(tx); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// updateSyncTimestamp updates the last_sync_timestamp in cache_metadata
func updateSyncTimestamp(tx *sql.Tx) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO cache_metadata (key, value)
		VALUES ('last_sync_timestamp', ?)
	`, now)
	if err != nil {
		return fmt.Errorf("failed to update sync timestamp: %w", err)
	}
	return nil
}
