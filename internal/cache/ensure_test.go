package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureCacheReady_CreatesMissingDatabase(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	if err := os.MkdirAll(pmPath, 0755); err != nil {
		t.Fatalf("Failed to create .pm directory: %v", err)
	}

	dbPath := filepath.Join(pmPath, ".cache.db")

	// Verify database doesn't exist initially
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatal("Database should not exist initially")
	}

	// Call EnsureCacheReady - should create the database
	if err := EnsureCacheReady(pmPath); err != nil {
		t.Fatalf("EnsureCacheReady failed: %v", err)
	}

	// Verify database was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database should exist after EnsureCacheReady")
	}

	// Verify we can open and query the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check that the tickets table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tickets'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Tickets table should exist: %v", err)
	}
	if tableName != "tickets" {
		t.Errorf("Expected table 'tickets', got '%s'", tableName)
	}
}

func TestEnsureCacheReady_IdempotentWithExistingDatabase(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	if err := os.MkdirAll(pmPath, 0755); err != nil {
		t.Fatalf("Failed to create .pm directory: %v", err)
	}

	// First call - creates database
	if err := EnsureCacheReady(pmPath); err != nil {
		t.Fatalf("First EnsureCacheReady failed: %v", err)
	}

	// Second call - should be idempotent (no error)
	if err := EnsureCacheReady(pmPath); err != nil {
		t.Fatalf("Second EnsureCacheReady failed (should be idempotent): %v", err)
	}

	// Third call - just to be sure
	if err := EnsureCacheReady(pmPath); err != nil {
		t.Fatalf("Third EnsureCacheReady failed: %v", err)
	}
}

func TestEnsureCacheReady_AppliesPendingMigrations(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	if err := os.MkdirAll(pmPath, 0755); err != nil {
		t.Fatalf("Failed to create .pm directory: %v", err)
	}

	dbPath := filepath.Join(pmPath, ".cache.db")

	// Call EnsureCacheReady - should create database with all migrations
	if err := EnsureCacheReady(pmPath); err != nil {
		t.Fatalf("EnsureCacheReady failed: %v", err)
	}

	// Open database and verify both migrations were applied
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check for tickets table (from migration 000001)
	var ticketsTable string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tickets'").Scan(&ticketsTable)
	if err != nil {
		t.Fatalf("Tickets table should exist: %v", err)
	}

	// Check for cache_metadata table (from migration 000002)
	var metadataTable string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='cache_metadata'").Scan(&metadataTable)
	if err != nil {
		t.Fatalf("Cache_metadata table should exist: %v", err)
	}
}
