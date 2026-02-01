package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetNextTicketNumber(t *testing.T) {
	tests := []struct {
		name            string
		existingTickets []string
		expectedNext    int
	}{
		{
			name:            "empty directory",
			existingTickets: []string{},
			expectedNext:    1,
		},
		{
			name:            "single ticket",
			existingTickets: []string{"TEST-1.md"},
			expectedNext:    2,
		},
		{
			name:            "sequential tickets",
			existingTickets: []string{"TEST-1.md", "TEST-2.md", "TEST-3.md"},
			expectedNext:    4,
		},
		{
			name:            "gap in sequence",
			existingTickets: []string{"TEST-1.md", "TEST-2.md", "TEST-5.md"},
			expectedNext:    6,
		},
		{
			name:            "non-sequential start",
			existingTickets: []string{"TEST-10.md", "TEST-20.md"},
			expectedNext:    21,
		},
		{
			name:            "mixed with other files",
			existingTickets: []string{"TEST-1.md", "README.md", "TEST-2.md", "notes.txt"},
			expectedNext:    3,
		},
		{
			name:            "different prefix ignored",
			existingTickets: []string{"TEST-1.md", "OTHER-999.md", "TEST-2.md"},
			expectedNext:    3,
		},
		{
			name:            "unsorted tickets",
			existingTickets: []string{"TEST-5.md", "TEST-1.md", "TEST-3.md"},
			expectedNext:    6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tempDir := t.TempDir()
			ticketsDir := filepath.Join(tempDir, ".pm", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatalf("Failed to create tickets directory: %v", err)
			}

			// Create test ticket files
			for _, filename := range tt.existingTickets {
				ticketPath := filepath.Join(ticketsDir, filename)
				content := `---
id: ` + filename[:len(filename)-3] + `
title: Test Ticket
type: task
status: todo
created_at: 2026-01-31T09:00:00Z
updated_at: 2026-01-31T09:00:00Z
---
Test content
`
				if err := os.WriteFile(ticketPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create ticket file: %v", err)
				}
			}

			// Test getNextTicketNumber
			nextNum := getNextTicketNumber(filepath.Dir(ticketsDir), "TEST")

			if nextNum != tt.expectedNext {
				t.Errorf("getNextTicketNumber() = %d, want %d", nextNum, tt.expectedNext)
			}
		})
	}
}

func TestGetNextTicketNumberInvalidDirectory(t *testing.T) {
	// getNextTicketNumber returns 1 if directory doesn't exist
	nextNum := getNextTicketNumber("/nonexistent/directory", "TEST")
	if nextNum != 1 {
		t.Errorf("getNextTicketNumber() with nonexistent directory = %d, want 1", nextNum)
	}
}

func TestGetNextTicketNumberInvalidFilenames(t *testing.T) {
	tempDir := t.TempDir()
	ticketsDir := filepath.Join(tempDir, ".pm", "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create tickets directory: %v", err)
	}

	// Create files with invalid formats (should be ignored)
	invalidFiles := []string{
		"TEST-.md",        // no number
		"TEST-abc.md",     // non-numeric
		"TEST-1-2.md",     // multiple dashes
		"-5.md",           // no prefix
		"TEST-1",          // no extension
		"TEST-1.txt",      // wrong extension
	}

	for _, filename := range invalidFiles {
		filePath := filepath.Join(ticketsDir, filename)
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Should return 1 since all files are invalid
	nextNum := getNextTicketNumber(filepath.Dir(ticketsDir), "TEST")

	if nextNum != 1 {
		t.Errorf("getNextTicketNumber() with invalid files = %d, want 1", nextNum)
	}
}
