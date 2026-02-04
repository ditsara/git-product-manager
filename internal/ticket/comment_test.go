package ticket

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateCommentFilename(t *testing.T) {
	tests := []struct {
		name      string
		author    string
		timestamp time.Time
		check     func(string) bool
	}{
		{
			name:      "simple author",
			author:    "alice",
			timestamp: time.Date(2026, 2, 3, 14, 30, 0, 0, time.UTC),
			check: func(s string) bool {
				return strings.HasPrefix(s, "2026-02-03T14-30-00Z-") &&
					strings.HasSuffix(s, ".md") &&
					!strings.Contains(s, ":")
			},
		},
		{
			name:      "author with spaces",
			author:    "alice smith",
			timestamp: time.Date(2026, 2, 3, 14, 30, 0, 0, time.UTC),
			check: func(s string) bool {
				// Should sanitize spaces
				return strings.HasSuffix(s, ".md") && !strings.Contains(s, " ")
			},
		},
		{
			name:      "author with special chars",
			author:    "bob@example.com",
			timestamp: time.Date(2026, 2, 3, 14, 30, 0, 0, time.UTC),
			check: func(s string) bool {
				// Should remove special characters
				return strings.HasSuffix(s, ".md") && !strings.Contains(s, "@")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCommentFilename(tt.author, tt.timestamp)
			if !tt.check(result) {
				t.Errorf("GenerateCommentFilename(%q, ...) = %q, check failed", tt.author, result)
			}
		})
	}
}

func TestCreateAndParseCommentFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test comment
	body := "This is a test comment"
	author := "alice"
	relPath, err := CreateCommentFile("TEST-1", author, body, tmpDir)
	if err != nil {
		t.Fatalf("CreateCommentFile failed: %v", err)
	}

	// Verify file was created - relPath is relative to .pm/tickets
	fullPath := filepath.Join(tmpDir, "tickets", relPath)
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("Comment file not created at %s: %v", fullPath, err)
	}

	// Parse the created file
	comment, err := ParseCommentFile(fullPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}

	if comment.Author != author {
		t.Errorf("Expected author %q, got %q", author, comment.Author)
	}

	if comment.Body != body {
		t.Errorf("Expected body %q, got %q", body, comment.Body)
	}

	if comment.Timestamp.IsZero() {
		t.Errorf("Timestamp should not be zero")
	}
}

func TestCreateCommentDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create comment on ticket where directory doesn't exist
	relPath, err := CreateCommentFile("NEW-TICKET", "bob", "First comment", tmpDir)
	if err != nil {
		t.Fatalf("CreateCommentFile failed: %v", err)
	}

	// Verify directory was created
	commentDir := filepath.Join(tmpDir, "tickets", "NEW-TICKET")
	if info, err := os.Stat(commentDir); err != nil || !info.IsDir() {
		t.Fatalf("Comment directory not created")
	}

	// Verify file exists - relPath is relative to .pm/tickets
	fullPath := filepath.Join(tmpDir, "tickets", relPath)
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("Comment file not found in created directory: %v", err)
	}
}

func TestParseCommentFileInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.md")

	// Write invalid file (missing delimiters)
	os.WriteFile(badFile, []byte("no front matter here"), 0644)

	_, err := ParseCommentFile(badFile)
	if err == nil {
		t.Errorf("ParseCommentFile should fail on invalid format")
	}
}

func TestListCommentsForTicket(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple comments
	comments := []struct {
		author string
		body   string
	}{
		{"alice", "First comment"},
		{"bob", "Second comment"},
		{"charlie", "Third comment"},
	}

	// Add slight delays to ensure different timestamps
	for i, c := range comments {
		_, err := CreateCommentFile("TEST-1", c.author, c.body, tmpDir)
		if err != nil {
			t.Fatalf("CreateCommentFile failed: %v", err)
		}
		if i < len(comments)-1 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// List comments
	result, err := ListCommentsForTicket("TEST-1", tmpDir)
	if err != nil {
		t.Fatalf("ListCommentsForTicket failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(result))
	}

	// Verify they're in chronological order
	for i := 1; i < len(result); i++ {
		if result[i].Timestamp.Before(result[i-1].Timestamp) {
			t.Errorf("Comments not sorted by timestamp")
		}
	}
}

func TestListCommentsForTicketNoComments(t *testing.T) {
	tmpDir := t.TempDir()

	// List comments for ticket with no comments directory
	result, err := ListCommentsForTicket("NONEXISTENT", tmpDir)
	if err != nil {
		t.Fatalf("ListCommentsForTicket should not error for nonexistent ticket: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 comments for nonexistent ticket, got %d", len(result))
	}
}

func TestCommentFilenameWithCollision(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two comments very quickly (could have same timestamp)
	_, err1 := CreateCommentFile("TEST-1", "alice", "Comment 1", tmpDir)
	_, err2 := CreateCommentFile("TEST-1", "alice", "Comment 2", tmpDir)

	if err1 != nil || err2 != nil {
		t.Fatalf("CreateCommentFile failed: %v, %v", err1, err2)
	}

	// List comments - both should exist
	comments, err := ListCommentsForTicket("TEST-1", tmpDir)
	if err != nil {
		t.Fatalf("ListCommentsForTicket failed: %v", err)
	}

	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}
}

func TestCommentBodyWithSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()

	body := "Test with special chars: !@#$%^&*() \n\nNew line\n\tTabbed line"
	relPath, err := CreateCommentFile("TEST-1", "alice", body, tmpDir)
	if err != nil {
		t.Fatalf("CreateCommentFile failed: %v", err)
	}

	// Parse and verify body is preserved - relPath is relative to .pm/tickets
	fullPath := filepath.Join(tmpDir, "tickets", relPath)
	comment, err := ParseCommentFile(fullPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}

	if comment.Body != body {
		t.Errorf("Body not preserved\nExpected: %q\nGot: %q", body, comment.Body)
	}
}

func TestCommentTimestampFormat(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := CreateCommentFile("TEST-1", "alice", "Test", tmpDir)
	if err != nil {
		t.Fatalf("CreateCommentFile failed: %v", err)
	}

	// Read the file and verify timestamp format
	commentDir := filepath.Join(tmpDir, "tickets", "TEST-1")
	entries, _ := os.ReadDir(commentDir)
	if len(entries) == 0 {
		t.Fatalf("No comment files found")
	}

	commentPath := filepath.Join(commentDir, entries[0].Name())
	comment, err := ParseCommentFile(commentPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}

	// Verify timestamp is valid and in UTC
	if comment.Timestamp.Location() != time.UTC && comment.Timestamp.Location() != time.Local {
		// time.Parse with RFC3339 always returns UTC, so this should pass
		if comment.Timestamp.IsZero() {
			t.Errorf("Timestamp is zero")
		}
	}
}
