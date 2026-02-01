package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestRunMigrations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Get the project root to find migrations
	projectRoot := getProjectRoot(t)
	migrationPath := filepath.Join(projectRoot, "migrations")

	// Test successful migration
	err := RunMigrations(dbPath, migrationPath)
	if err != nil {
		t.Fatalf("RunMigrations() unexpected error = %v", err)
	}

	// Verify database was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("runMigrations() did not create database file")
	}

	// Verify database is accessible
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database after migration: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Errorf("runMigrations() created database is not accessible: %v", err)
	}

	// Verify tickets table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tickets'").Scan(&tableName)
	if err != nil {
		t.Errorf("runMigrations() did not create tickets table: %v", err)
	}
	if tableName != "tickets" {
		t.Errorf("runMigrations() table name = %v, want 'tickets'", tableName)
	}

	// Verify table structure
	rows, err := db.Query("PRAGMA table_info(tickets)")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]bool{
		"id":         false,
		"title":      false,
		"type":       false,
		"status":     false,
		"priority":   false,
		"assignee":   false,
		"parent":     false,
		"created_at": false,
		"updated_at": false,
		"body":       false,
	}

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString

		err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
		}
	}

	// Verify all expected columns were found
	for col, found := range expectedColumns {
		if !found {
			t.Errorf("runMigrations() tickets table missing column: %s", col)
		}
	}
}

func TestRunMigrationsIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Get the project root to find migrations
	projectRoot := getProjectRoot(t)
	migrationPath := filepath.Join(projectRoot, "migrations")

	// Run migrations twice
	err := RunMigrations(dbPath, migrationPath)
	if err != nil {
		t.Fatalf("RunMigrations() first run unexpected error = %v", err)
	}

	err = RunMigrations(dbPath, migrationPath)
	if err != nil {
		t.Fatalf("RunMigrations() second run unexpected error = %v", err)
	}

	// Verify database is still functional
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Errorf("runMigrations() database not accessible after second run: %v", err)
	}
}

func TestRunMigrationsInvalidPath(t *testing.T) {
	// Try to create database in nonexistent directory
	projectRoot := getProjectRoot(t)
	migrationPath := filepath.Join(projectRoot, "migrations")
	
	err := RunMigrations("/nonexistent/invalid/path/test.db", migrationPath)
	if err == nil {
		t.Error("RunMigrations() expected error for invalid path, got nil")
	}
}

// Helper function to get project root directory
func getProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current file's directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}
