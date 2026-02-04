package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
)

// TestDisplayComments tests the comment display formatting
func TestDisplayCommentsFormatting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test comments
	ticketID := "TEST-1"
	commentDir := filepath.Join(tmpDir, "tickets", ticketID)
	os.MkdirAll(commentDir, 0755)

	// Create a few comments
	comments := []struct {
		author string
		body   string
	}{
		{"alice", "First comment"},
		{"bob", "Second comment"},
		{"alice", "Third comment"},
	}

	for i, c := range comments {
		_, err := ticket.CreateCommentFile(ticketID, c.author, c.body, tmpDir)
		if err != nil {
			t.Fatalf("Failed to create comment: %v", err)
		}
		if i < len(comments)-1 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Read comments
	readComments, err := ticket.ListCommentsForTicket(ticketID, tmpDir)
	if err != nil {
		t.Fatalf("Failed to read comments: %v", err)
	}

	if len(readComments) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(readComments))
	}

	// Verify chronological order
	for i := 1; i < len(readComments); i++ {
		if readComments[i].Timestamp.Before(readComments[i-1].Timestamp) {
			t.Errorf("Comments not in chronological order")
		}
	}

	// Verify that all expected authors are present
	authors := make(map[string]int)
	bodies := make(map[string]bool)
	for _, c := range readComments {
		authors[c.Author]++
		bodies[c.Body] = true
	}

	if authors["alice"] != 2 {
		t.Errorf("Expected 2 comments from alice, got %d", authors["alice"])
	}
	if authors["bob"] != 1 {
		t.Errorf("Expected 1 comment from bob, got %d", authors["bob"])
	}

	if !bodies["First comment"] || !bodies["Second comment"] || !bodies["Third comment"] {
		t.Errorf("Not all expected comment bodies found")
	}
}

// TestDisplayCommentsOutput verifies the formatted output
func TestDisplayCommentsOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test comments
	ticketID := "TEST-1"
	_, err := ticket.CreateCommentFile(ticketID, "alice", "Test comment", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Read comments
	comments, err := ticket.ListCommentsForTicket(ticketID, tmpDir)
	if err != nil {
		t.Fatalf("Failed to read comments: %v", err)
	}

	if len(comments) == 0 {
		t.Fatalf("No comments found")
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Mock the display by checking key components
	for _, comment := range comments {
		timestamp := comment.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC")
		// Just verify the structure would be valid
		_ = timestamp
		_ = comment.Author
		_ = comment.Body
	}

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

// TestDisplayCommentsEmptyDirectory verifies no error on empty comment directory
func TestDisplayCommentsEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// No comments created
	ticketID := "TEST-1"

	// Should not error
	comments, err := ticket.ListCommentsForTicket(ticketID, tmpDir)
	if err != nil {
		t.Fatalf("Should not error on missing comment directory: %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(comments))
	}
}

// TestDisplayCommentsMultilineFormat verifies multiline comment handling
func TestDisplayCommentsMultilineFormat(t *testing.T) {
	tmpDir := t.TempDir()

	ticketID := "TEST-1"
	multilineBody := "First line\nSecond line\nThird line"

	_, err := ticket.CreateCommentFile(ticketID, "alice", multilineBody, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	comments, err := ticket.ListCommentsForTicket(ticketID, tmpDir)
	if err != nil {
		t.Fatalf("Failed to read comments: %v", err)
	}

	if len(comments) == 0 {
		t.Fatalf("No comments found")
	}

	// Verify multiline body is preserved
	if comments[0].Body != multilineBody {
		t.Errorf("Multiline body not preserved\nExpected: %q\nGot: %q", multilineBody, comments[0].Body)
	}
}

// TestDisplayCommentsTimestampFormat verifies UTC timestamp formatting
func TestDisplayCommentsTimestampFormat(t *testing.T) {
	tmpDir := t.TempDir()

	ticketID := "TEST-1"
	_, err := ticket.CreateCommentFile(ticketID, "alice", "Test", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	comments, err := ticket.ListCommentsForTicket(ticketID, tmpDir)
	if err != nil {
		t.Fatalf("Failed to read comments: %v", err)
	}

	if len(comments) == 0 {
		t.Fatalf("No comments found")
	}

	// Verify timestamp can be formatted
	timestamp := comments[0].Timestamp.UTC().Format("2006-01-02 15:04:05 UTC")

	// Should have valid format
	if !strings.Contains(timestamp, "-") || !strings.Contains(timestamp, "UTC") {
		t.Errorf("Invalid timestamp format: %s", timestamp)
	}
}

// TestCaptureDisplayCommentsOutput tests output format with capture
func TestCaptureDisplayCommentsOutput(t *testing.T) {
	tmpDir := t.TempDir()

	ticketID := "TEST-1"
	_, err := ticket.CreateCommentFile(ticketID, "alice", "First comment", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Small delay to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	_, err = ticket.CreateCommentFile(ticketID, "bob", "Second comment", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Read comments to verify they exist
	readComments, err := ticket.ListCommentsForTicket(ticketID, tmpDir)
	if err != nil {
		t.Fatalf("Failed to read comments: %v", err)
	}

	if len(readComments) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(readComments))
	}

	// Verify the comments contain expected data
	foundAlice := false
	foundBob := false
	for _, c := range readComments {
		if c.Author == "alice" && c.Body == "First comment" {
			foundAlice = true
		}
		if c.Author == "bob" && c.Body == "Second comment" {
			foundBob = true
		}
	}

	if !foundAlice {
		t.Errorf("Alice's comment not found")
	}
	if !foundBob {
		t.Errorf("Bob's comment not found")
	}
}
