package main

import (
	"strings"
	"testing"
	"time"
)

// TestShowWithComments tests showing ticket with comments
func TestShowWithComments(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "SHOW")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket with Comments")

	// Add multiple comments
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "First comment from alice", "--author", "alice")
	time.Sleep(10 * time.Millisecond)
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "Second comment from bob", "--author", "bob")
	time.Sleep(10 * time.Millisecond)
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "Third comment from alice", "--author", "alice")

	// Show ticket - should display comments
	output, err := runPM(t, pmBinary, workspace, "show", "SHOW-1")
	if err != nil {
		t.Fatalf("pm show failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected elements
	if !strings.Contains(output, "Comments (3):") {
		t.Errorf("Output missing comment count: %s", output)
	}

	if !strings.Contains(output, "@alice") {
		t.Errorf("Output missing alice author: %s", output)
	}

	if !strings.Contains(output, "@bob") {
		t.Errorf("Output missing bob author: %s", output)
	}

	if !strings.Contains(output, "First comment from alice") {
		t.Errorf("Output missing first comment: %s", output)
	}

	if !strings.Contains(output, "Second comment from bob") {
		t.Errorf("Output missing second comment: %s", output)
	}

	if !strings.Contains(output, "Third comment from alice") {
		t.Errorf("Output missing third comment: %s", output)
	}

	if !strings.Contains(output, "━━━━━━━━━━━━━━━") {
		t.Errorf("Output missing separator: %s", output)
	}
}

// TestShowNoComments tests showing a ticket without comments
func TestShowNoComments(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "SHOW")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket No Comments")

	// Show ticket without adding comments
	output, err := runPM(t, pmBinary, workspace, "show", "SHOW-1")
	if err != nil {
		t.Fatalf("pm show failed: %v\nOutput: %s", err, output)
	}

	// Should not have comments section (check for the comments header specifically)
	if strings.Contains(output, "Comments (") || strings.Contains(output, "━━━━━━━━━━━━━━━") {
		t.Errorf("Output should not have comments section: %s", output)
	}

	// But should still show ticket content
	if !strings.Contains(output, "Test Ticket No Comments") {
		t.Errorf("Output missing ticket title: %s", output)
	}
}

// TestShowWithNoCommentsFlag tests --no-comments flag
func TestShowWithNoCommentsFlag(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "SHOW")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket with Flag")

	// Add a comment
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "This comment should be hidden", "--author", "alice")

	// Show ticket with --no-comments flag
	output, err := runPM(t, pmBinary, workspace, "show", "SHOW-1", "--no-comments")
	if err != nil {
		t.Fatalf("pm show --no-comments failed: %v\nOutput: %s", err, output)
	}

	// Should not have comments section
	if strings.Contains(output, "Comments") {
		t.Errorf("Output should not have comments section with --no-comments: %s", output)
	}

	if strings.Contains(output, "This comment should be hidden") {
		t.Errorf("Comment should not be visible with --no-comments: %s", output)
	}

	// But should still show ticket content
	if !strings.Contains(output, "Test Ticket with Flag") {
		t.Errorf("Output missing ticket title: %s", output)
	}
}

// TestShowCommentChronologicalOrder tests that comments are displayed in order
func TestShowCommentChronologicalOrder(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "SHOW")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Chronological")

	// Add comments in specific order
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "Comment 1", "--author", "alice")
	time.Sleep(10 * time.Millisecond)
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "Comment 2", "--author", "bob")
	time.Sleep(10 * time.Millisecond)
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "Comment 3", "--author", "charlie")

	// Show ticket
	output, err := runPM(t, pmBinary, workspace, "show", "SHOW-1")
	if err != nil {
		t.Fatalf("pm show failed: %v\nOutput: %s", err, output)
	}

	// Verify order: Comment 1 should appear before Comment 2, which should appear before Comment 3
	pos1 := strings.Index(output, "Comment 1")
	pos2 := strings.Index(output, "Comment 2")
	pos3 := strings.Index(output, "Comment 3")

	if pos1 == -1 || pos2 == -1 || pos3 == -1 {
		t.Fatalf("Comments not found in output: %s", output)
	}

	if pos1 > pos2 {
		t.Errorf("Comment 1 should appear before Comment 2")
	}

	if pos2 > pos3 {
		t.Errorf("Comment 2 should appear before Comment 3")
	}
}

// TestShowCaseInsensitiveWithComments tests case-insensitive show with comments
func TestShowCaseInsensitiveWithComments(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "SHOW")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Case Insensitive Show")

	// Add a comment
	runPM(t, pmBinary, workspace, "comment", "SHOW-1", "-m", "Test comment", "--author", "alice")

	// Show with different cases
	for _, id := range []string{"show-1", "SHOW-1", "Show-1"} {
		output, err := runPM(t, pmBinary, workspace, "show", id)
		if err != nil {
			t.Fatalf("pm show %s failed: %v\nOutput: %s", id, err, output)
		}

		// Should show comments regardless of case
		if !strings.Contains(output, "Comments (1):") {
			t.Errorf("Expected comments with ID %q: %s", id, output)
		}

		if !strings.Contains(output, "Test comment") {
			t.Errorf("Expected comment content with ID %q: %s", id, output)
		}
	}
}
