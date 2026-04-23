package cache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsRecoverableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"no such table", errors.New("no such table: tickets"), true},
		{"database is locked", errors.New("database is locked"), true},
		{"file is not a database", errors.New("file is not a database"), true},
		{"unable to open database file", errors.New("unable to open database file"), true},
		{"unrelated error", errors.New("permission denied"), false},
		{"connection refused", errors.New("connection refused"), false},
		// Case insensitive
		{"uppercase variant", errors.New("NO SUCH TABLE: foo"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRecoverableError(tt.err)
			if got != tt.want {
				t.Errorf("isRecoverableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRecoverFromError_NilError(t *testing.T) {
	err := RecoverFromError(nil, "/some/path")
	if err != nil {
		t.Errorf("RecoverFromError(nil, ...) = %v, want nil", err)
	}
}

func TestRecoverFromError_NonRecoverableError(t *testing.T) {
	original := errors.New("permission denied: not a db error")
	err := RecoverFromError(original, "/some/path")
	if err != original {
		t.Errorf("RecoverFromError(non-recoverable) should return original error, got %v", err)
	}
}

func TestRecoverFromError_CorruptedDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	ticketsPath := filepath.Join(pmPath, "tickets")
	if err := os.MkdirAll(ticketsPath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Write a corrupted (non-SQLite) database file
	dbPath := filepath.Join(pmPath, ".cache.db")
	if err := os.WriteFile(dbPath, []byte("not a sqlite database"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted db: %v", err)
	}

	// Simulate the error returned when SQLite tries to open a corrupted file
	corruptErr := errors.New("file is not a database")

	err := RecoverFromError(corruptErr, pmPath)
	if err != nil {
		t.Fatalf("RecoverFromError failed: %v", err)
	}

	// Database should now exist and be valid
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database should exist after recovery")
	}

	// Should be a valid SQLite file (open and query succeeds)
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open recovered database: %v", err)
	}
	defer db.Close()

	var tableName string
	queryErr := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tickets'").Scan(&tableName)
	if queryErr != nil {
		t.Fatalf("Tickets table should exist after recovery: %v", queryErr)
	}
}

func TestRecoverFromError_MissingDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	if err := os.MkdirAll(pmPath, 0755); err != nil {
		t.Fatalf("Failed to create .pm dir: %v", err)
	}

	// Simulate a "no such table" error (DB exists but has no schema)
	noTableErr := errors.New("no such table: tickets")

	err := RecoverFromError(noTableErr, pmPath)
	if err != nil {
		t.Fatalf("RecoverFromError failed: %v", err)
	}

	dbPath := filepath.Join(pmPath, ".cache.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database should exist after recovery")
	}
}

func TestEnsureCacheReady_AutoRecoveryFromCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	if err := os.MkdirAll(pmPath, 0755); err != nil {
		t.Fatalf("Failed to create .pm dir: %v", err)
	}

	// Write a corrupted database file
	dbPath := filepath.Join(pmPath, ".cache.db")
	if err := os.WriteFile(dbPath, []byte("this is garbage, not sqlite"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted db: %v", err)
	}

	// EnsureCacheReady should detect corruption and auto-recover
	if err := EnsureCacheReady(pmPath); err != nil {
		t.Fatalf("EnsureCacheReady should auto-recover from corruption, got: %v", err)
	}

	// DB should be valid after recovery
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open recovered database: %v", err)
	}
	defer db.Close()

	var tableName string
	if err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tickets'").Scan(&tableName); err != nil {
		t.Fatalf("Tickets table should exist after recovery: %v", err)
	}
	if tableName != "tickets" {
		t.Errorf("Expected table 'tickets', got %q", tableName)
	}
}

func TestRecoverFromError_PrintsMessages(t *testing.T) {
	// Redirect stderr to capture output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	if err := os.MkdirAll(pmPath, 0755); err != nil {
		os.Stderr = oldStderr
		t.Fatalf("Failed to create .pm dir: %v", err)
	}

	corruptErr := errors.New("file is not a database")
	RecoverFromError(corruptErr, pmPath) //nolint

	w.Close()
	os.Stderr = oldStderr

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "corrupted") && !strings.Contains(output, "rebuilding") {
		t.Errorf("Expected warning message about corruption, got: %q", output)
	}
	if !strings.Contains(output, "rebuilt") && !strings.Contains(output, "success") {
		t.Errorf("Expected success message after rebuild, got: %q", output)
	}
}
