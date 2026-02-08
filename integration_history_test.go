package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestHistorySingleChange tests pm history with one status change
func TestHistorySingleChange(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "HIST")
	initGitRepo(t, workspace)

	output, err := runPM(t, pmBinary, workspace, "new", "History test ticket")
	if err != nil {
		t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "HIST-1") {
		t.Fatalf("Expected HIST-1 in output, got: %s", output)
	}

	// Commit initial ticket
	cmd := exec.Command("git", "add", ".pm/tickets/HIST-1.md")
	cmd.Dir = workspace
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\nOutput: %s", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", "Create ticket")
	cmd.Dir = workspace
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\nOutput: %s", err, out)
	}

	// Move and commit status change
	if _, err := runPM(t, pmBinary, workspace, "move", "HIST-1", "todo"); err != nil {
		t.Fatalf("pm move failed: %v", err)
	}
	cmd = exec.Command("git", "add", ".pm/tickets/HIST-1.md")
	cmd.Dir = workspace
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\nOutput: %s", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", "Move to todo")
	cmd.Dir = workspace
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\nOutput: %s", err, out)
	}

	history, err := runPM(t, pmBinary, workspace, "history", "HIST-1")
	if err != nil {
		t.Fatalf("pm history failed: %v\nOutput: %s", err, history)
	}

	if !strings.Contains(history, "Created (status: backlog)") {
		t.Errorf("Missing creation entry: %s", history)
	}
	if !strings.Contains(history, "backlog → todo") {
		t.Errorf("Missing status transition: %s", history)
	}
	if !strings.Contains(history, "commit: Move to todo") {
		t.Errorf("Missing commit message: %s", history)
	}
}

// TestHistoryNoGitRepo tests pm history when not in a git repo
func TestHistoryNoGitRepo(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "HIST")
	if _, err := runPM(t, pmBinary, workspace, "new", "History test ticket"); err != nil {
		t.Fatalf("pm new failed: %v", err)
	}

	output, err := runPM(t, pmBinary, workspace, "history", "HIST-1")
	if err == nil {
		t.Fatalf("pm history should fail outside git repo")
	}
	if !strings.Contains(output, "git not available") {
		t.Errorf("Expected git error, got: %s", output)
	}
}

// TestHistoryTicketNotCommitted tests pm history when ticket not committed
func TestHistoryTicketNotCommitted(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "HIST")
	initGitRepo(t, workspace)
	if _, err := runPM(t, pmBinary, workspace, "new", "History test ticket"); err != nil {
		t.Fatalf("pm new failed: %v", err)
	}

	output, err := runPM(t, pmBinary, workspace, "history", "HIST-1")
	if err == nil {
		t.Fatalf("pm history should fail when ticket not committed")
	}
	if !strings.Contains(output, "ticket not in git history") {
		t.Errorf("Expected ticket history error, got: %s", output)
	}
}
