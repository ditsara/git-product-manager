package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditDescriptionFlag(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()

	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "EDIT")

	output, err := runPM(t, pmBinary, workspace, "new", "Original title")
	if err != nil {
		t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
	}

	ticketID := "EDIT-1"
	newBody := "# Description\nUpdated from flag"

	cmd := exec.Command(pmBinary, "edit", ticketID, "--description", newBody)
	cmd.Dir = workspace
	cmd.Env = append(os.Environ(), "EDITOR=/definitely-not-a-real-editor")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pm edit --description failed: %v\nOutput: %s", err, outputBytes)
	}

	ticketPath := filepath.Join(workspace, ".pm", "tickets", ticketID+".md")
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("Failed to read ticket file: %v", err)
	}

	if got := strings.TrimSpace(strings.SplitN(string(content), "---", 3)[2]); got != newBody {
		t.Fatalf("expected body %q, got %q", newBody, got)
	}
}

func TestEditFieldAndDescriptionTogether(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()

	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "EDIT")

	output, err := runPM(t, pmBinary, workspace, "new", "Original title")
	if err != nil {
		t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
	}

	ticketID := "EDIT-1"
	newBody := "# Description\nCombined update"

	output, err = runPM(t, pmBinary, workspace, "edit", ticketID, "--field", "priority=high", "--description", newBody)
	if err != nil {
		t.Fatalf("pm edit combined update failed: %v\nOutput: %s", err, output)
	}

	ticketPath := filepath.Join(workspace, ".pm", "tickets", ticketID+".md")
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("Failed to read ticket file: %v", err)
	}

	if !strings.Contains(string(content), "priority: high") {
		t.Fatalf("expected updated priority in ticket file, got:\n%s", content)
	}

	if got := strings.TrimSpace(strings.SplitN(string(content), "---", 3)[2]); got != newBody {
		t.Fatalf("expected body %q, got %q", newBody, got)
	}
}
