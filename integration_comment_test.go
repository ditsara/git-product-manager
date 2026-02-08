package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
)

// TestCommentDirectMode tests the pm comment command with -m flag
func TestCommentDirectMode(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	// Initialize a test workspace
	initWorkspace(t, pmBinary, workspace, "COMMENT")

	// Create a test ticket
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket for Comments")

	// Add a comment via direct mode
	output, err := runPM(t, pmBinary, workspace, "comment", "COMMENT-1", "-m", "This is a test comment")
	if err != nil {
		t.Fatalf("pm comment failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "✓ Comment added") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify comment file was created
	commentFile := filepath.Join(workspace, ".pm", "tickets", "COMMENT-1")
	entries, err := os.ReadDir(commentFile)
	if err != nil {
		t.Fatalf("Failed to read comment directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatalf("No comment files found")
	}

	// Read the comment file and verify content
	commentPath := filepath.Join(commentFile, entries[0].Name())
	content, err := os.ReadFile(commentPath)
	if err != nil {
		t.Fatalf("Failed to read comment file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "This is a test comment") {
		t.Errorf("Comment body not found in file")
	}

	if !strings.Contains(contentStr, "author:") {
		t.Errorf("Author metadata not found in comment file")
	}

	if !strings.Contains(contentStr, "created_at:") {
		t.Errorf("created_at metadata not found in comment file")
	}
	if !strings.Contains(contentStr, "updated_at:") {
		t.Errorf("updated_at metadata not found in comment file")
	}
}

// TestCommentWithCustomAuthor tests overriding the author
func TestCommentWithCustomAuthor(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "COMMENT")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket")

	// Add comment with custom author
	output, err := runPM(t, pmBinary, workspace, "comment", "COMMENT-1", "-m", "Custom author comment", "--author", "bob")
	if err != nil {
		t.Fatalf("pm comment with --author failed: %v\nOutput: %s", err, output)
	}

	// Verify author in comment file
	commentFile := filepath.Join(workspace, ".pm", "tickets", "COMMENT-1")
	entries, _ := os.ReadDir(commentFile)
	commentPath := filepath.Join(commentFile, entries[0].Name())
	content, _ := os.ReadFile(commentPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "author: bob") {
		t.Errorf("Custom author 'bob' not found in comment metadata")
	}
	if !strings.Contains(contentStr, "created_at:") {
		t.Errorf("created_at metadata not found in comment file")
	}
	if !strings.Contains(contentStr, "updated_at:") {
		t.Errorf("updated_at metadata not found in comment file")
	}

}

func TestCommentOnNonexistentTicket(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "COMMENT")

	// Try to add comment on ticket that doesn't exist
	output, err := runPM(t, pmBinary, workspace, "comment", "COMMENT-999", "-m", "This should fail")
	if err == nil {
		t.Fatalf("pm comment should fail for nonexistent ticket")
	}

	if !strings.Contains(output, "not found") {
		t.Errorf("Expected 'not found' error message, got: %s", output)
	}
}

// TestMultipleComments tests adding multiple comments to the same ticket
func TestMultipleComments(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "COMMENT")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket")

	// Add multiple comments
	for i := 1; i <= 3; i++ {
		_, err := runPM(t, pmBinary, workspace, "comment", "COMMENT-1", "-m", "Comment "+string(rune(i)))
		if err != nil {
			t.Fatalf("Failed to add comment %d: %v", i, err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Verify all comments exist
	commentDir := filepath.Join(workspace, ".pm", "tickets", "COMMENT-1")
	entries, err := os.ReadDir(commentDir)
	if err != nil {
		t.Fatalf("Failed to read comment directory: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(entries))
	}
}

// TestCommentEmptyMessage tests that empty comments are rejected
func TestCommentEmptyMessage(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "COMMENT")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket")

	// Try to add empty comment
	output, err := runPM(t, pmBinary, workspace, "comment", "COMMENT-1", "-m", "   ")
	if err == nil {
		t.Fatalf("pm comment should fail for empty message")
	}

	if !strings.Contains(output, "empty") {
		t.Errorf("Expected 'empty' error message, got: %s", output)
	}
}

// TestCommentFilenameFormat tests that comment filenames follow the correct format
func TestCommentFilenameFormat(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "COMMENT")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket")

	// Add comment
	runPM(t, pmBinary, workspace, "comment", "COMMENT-1", "-m", "Test", "--author", "alice")

	// Verify filename format: ISO8601-timestamp-author.md with hyphens instead of colons
	commentDir := filepath.Join(workspace, ".pm", "tickets", "COMMENT-1")
	entries, _ := os.ReadDir(commentDir)
	filename := entries[0].Name()

	// Should match pattern: YYYY-MM-DDTHH-MM-SSZ-author.md
	if !strings.HasSuffix(filename, ".md") {
		t.Errorf("Comment filename should end with .md: %s", filename)
	}

	if !strings.Contains(filename, "T") {
		t.Errorf("Filename should contain 'T' date separator: %s", filename)
	}

	if !strings.Contains(filename, "Z-alice") {
		t.Errorf("Filename should contain 'Z-alice' (UTC marker and author): %s", filename)
	}

	// Should NOT contain colons (they're replaced with hyphens)
	if strings.Count(filename, ":") > 0 {
		t.Errorf("Filename should not contain colons: %s", filename)
	}
}

// TestCommentCaseInsensitive tests that pm comment works with various case combinations
func TestCommentCaseInsensitive(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "CASE")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Test Ticket")

	testCases := []string{
		"case-1", // lowercase
		"CASE-1", // uppercase
		"Case-1", // mixed case
		"cAsE-1", // random case
	}

	for i, ticketID := range testCases {
		output, err := runPM(t, pmBinary, workspace, "comment", ticketID, "-m", fmt.Sprintf("Comment %d", i+1))
		if err != nil {
			t.Fatalf("pm comment with %q failed: %v\nOutput: %s", ticketID, err, output)
		}

		if !strings.Contains(output, "✓ Comment added") {
			t.Errorf("Expected success message for %q, got: %s", ticketID, output)
		}
	}

	// Verify all comments are in the same directory (CASE-1, the canonical form)
	commentDir := filepath.Join(workspace, ".pm", "tickets", "CASE-1")
	entries, _ := os.ReadDir(commentDir)

	if len(entries) != 4 {
		t.Errorf("Expected 4 comments in CASE-1 directory, got %d", len(entries))
	}
}

// TestCommentAmendDirectMode tests updating a comment with --amend and -m
func TestCommentAmendDirectMode(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "AMEND")
	runPM(t, pmBinary, workspace, "new", "Test Ticket")

	// Create initial comment
	_, err := runPM(t, pmBinary, workspace, "comment", "AMEND-1", "-m", "Original comment", "--author", "alice")
	if err != nil {
		t.Fatalf("pm comment failed: %v", err)
	}

	commentDir := filepath.Join(workspace, ".pm", "tickets", "AMEND-1")
	entries, err := os.ReadDir(commentDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("Expected 1 comment file, got %d", len(entries))
	}

	commentPath := filepath.Join(commentDir, entries[0].Name())
	before, err := ticket.ParseCommentFile(commentPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}

	// Ensure updated_at changes
	time.Sleep(1100 * time.Millisecond)

	// Amend comment
	_, err = runPM(t, pmBinary, workspace, "comment", "AMEND-1", "--amend", "-m", "Updated comment", "--author", "alice")
	if err != nil {
		t.Fatalf("pm comment --amend failed: %v", err)
	}

	after, err := ticket.ParseCommentFile(commentPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}

	if after.Body != "Updated comment" {
		t.Errorf("Expected updated body, got %q", after.Body)
	}
	if !after.CreatedAt.Equal(before.CreatedAt) {
		t.Errorf("created_at should not change")
	}
	if !after.UpdatedAt.After(before.UpdatedAt) {
		t.Errorf("updated_at should be newer after amend")
	}
}

// TestCommentAmendByTimestamp tests selecting a comment by timestamp
func TestCommentAmendByTimestamp(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "AMEND")
	runPM(t, pmBinary, workspace, "new", "Test Ticket")

	// Create two comments
	_, err := runPM(t, pmBinary, workspace, "comment", "AMEND-1", "-m", "First comment", "--author", "bob")
	if err != nil {
		t.Fatalf("pm comment failed: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	_, err = runPM(t, pmBinary, workspace, "comment", "AMEND-1", "-m", "Second comment", "--author", "bob")
	if err != nil {
		t.Fatalf("pm comment failed: %v", err)
	}

	commentDir := filepath.Join(workspace, ".pm", "tickets", "AMEND-1")
	entries, err := os.ReadDir(commentDir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("Expected 2 comment files, got %d", len(entries))
	}

	firstPath := filepath.Join(commentDir, entries[0].Name())
	secondPath := filepath.Join(commentDir, entries[1].Name())
	firstComment, err := ticket.ParseCommentFile(firstPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}
	secondComment, err := ticket.ParseCommentFile(secondPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}

	// Identify the oldest comment by created_at
	oldest := firstComment
	oldestPath := firstPath
	if secondComment.CreatedAt.Before(firstComment.CreatedAt) {
		oldest = secondComment
		oldestPath = secondPath
	}

	// Amend the oldest comment by timestamp
	_, err = runPM(t, pmBinary, workspace, "comment", "AMEND-1", "--amend", "--timestamp", oldest.CreatedAt.Format(time.RFC3339), "-m", "Updated oldest", "--author", "bob")
	if err != nil {
		t.Fatalf("pm comment --amend --timestamp failed: %v", err)
	}

	updatedOldest, err := ticket.ParseCommentFile(oldestPath)
	if err != nil {
		t.Fatalf("ParseCommentFile failed: %v", err)
	}
	if updatedOldest.Body != "Updated oldest" {
		t.Errorf("Expected oldest comment to be updated")
	}
}
