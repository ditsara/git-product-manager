package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindTicketByID_ExactMatch(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, ".pm", "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create tickets directory: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create test tickets
	testTickets := []string{"TEST-1.md", "TEST-10.md", "TEST-100.md"}
	for _, ticket := range testTickets {
		path := filepath.Join(ticketsDir, ticket)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test ticket: %v", err)
		}
	}

	// Test exact match for TEST-1 (should not match TEST-10 or TEST-100)
	result := findTicketByID("TEST-1")
	expected := filepath.Join(".pm", "tickets", "TEST-1.md")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFindTicketByID_PrefixMatch(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, ".pm", "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create tickets directory: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create only TEST-10.md (no TEST-1.md)
	path := filepath.Join(ticketsDir, "TEST-10.md")
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test ticket: %v", err)
	}

	// Test prefix match: TEST-1 should match TEST-10 when no exact match exists
	result := findTicketByID("TEST-1")
	expected := filepath.Join(".pm", "tickets", "TEST-10.md")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFindTicketByID_AmbiguousMatch(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, ".pm", "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create tickets directory: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create multiple tickets that start with TEST-1
	// But no exact TEST-1.md
	testTickets := []string{"TEST-10.md", "TEST-11.md", "TEST-12.md"}
	for _, ticket := range testTickets {
		path := filepath.Join(ticketsDir, ticket)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test ticket: %v", err)
		}
	}

	// This should call os.Exit(1) due to ambiguity
	// We can't easily test os.Exit in a unit test, so we'll skip this
	// In a real scenario, we'd refactor to return an error instead of calling os.Exit
	t.Skip("Cannot test os.Exit in unit test - function should be refactored to return errors")
}

func TestFindTicketByID_NotFound(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, ".pm", "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create tickets directory: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create a ticket that doesn't match
	path := filepath.Join(ticketsDir, "OTHER-1.md")
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test ticket: %v", err)
	}

	// This should call os.Exit(1) due to not found
	// We can't easily test os.Exit in a unit test
	t.Skip("Cannot test os.Exit in unit test - function should be refactored to return errors")
}
