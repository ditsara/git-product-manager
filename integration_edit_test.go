package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestEditTouchAlone(t *testing.T) {
pmBinary := buildPMBinary(t)
workspace := t.TempDir()

initGitRepo(t, workspace)
initWorkspace(t, pmBinary, workspace, "EDIT")

_, err := runPM(t, pmBinary, workspace, "new", "Touch test ticket")
if err != nil {
t.Fatalf("pm new failed: %v", err)
}

ticketID := "EDIT-1"
ticketPath := filepath.Join(workspace, ".pm", "tickets", ticketID+".md")

originalContent, err := os.ReadFile(ticketPath)
if err != nil {
t.Fatalf("Failed to read ticket: %v", err)
}
originalUpdatedAt := extractYAMLField(string(originalContent), "updated_at")
originalTitle := extractYAMLField(string(originalContent), "title")

// Ensure time advances before touch
time.Sleep(1100 * time.Millisecond)

output, err := runPM(t, pmBinary, workspace, "edit", ticketID, "--touch")
if err != nil {
t.Fatalf("pm edit --touch failed: %v\nOutput: %s", err, output)
}
if !strings.Contains(output, "Touched") {
t.Errorf("expected 'Touched' in output, got: %s", output)
}

newContent, err := os.ReadFile(ticketPath)
if err != nil {
t.Fatalf("Failed to read ticket after touch: %v", err)
}

newUpdatedAt := extractYAMLField(string(newContent), "updated_at")
if newUpdatedAt == originalUpdatedAt {
t.Errorf("expected updated_at to change after --touch, stayed %q", originalUpdatedAt)
}
if newTitle := extractYAMLField(string(newContent), "title"); newTitle != originalTitle {
t.Errorf("--touch should not change title: before %q, after %q", originalTitle, newTitle)
}
}

func TestEditTouchWithField(t *testing.T) {
pmBinary := buildPMBinary(t)
workspace := t.TempDir()

initGitRepo(t, workspace)
initWorkspace(t, pmBinary, workspace, "EDIT")

_, err := runPM(t, pmBinary, workspace, "new", "Touch+field test")
if err != nil {
t.Fatalf("pm new failed: %v", err)
}

ticketID := "EDIT-1"
ticketPath := filepath.Join(workspace, ".pm", "tickets", ticketID+".md")

originalContent, err := os.ReadFile(ticketPath)
if err != nil {
t.Fatalf("Failed to read ticket: %v", err)
}
originalUpdatedAt := extractYAMLField(string(originalContent), "updated_at")

time.Sleep(1100 * time.Millisecond)

output, err := runPM(t, pmBinary, workspace, "edit", ticketID, "--touch", "--field", "priority=high")
if err != nil {
t.Fatalf("pm edit --touch --field failed: %v\nOutput: %s", err, output)
}

newContent, err := os.ReadFile(ticketPath)
if err != nil {
t.Fatalf("Failed to read ticket after touch+field: %v", err)
}
if !strings.Contains(string(newContent), "priority: high") {
t.Errorf("expected priority=high in ticket, got:\n%s", newContent)
}
if newUpdatedAt := extractYAMLField(string(newContent), "updated_at"); newUpdatedAt == originalUpdatedAt {
t.Errorf("expected updated_at to change, stayed %q", originalUpdatedAt)
}
}

// extractYAMLField is a simple line-scanner for front-matter fields in tests.
// It returns the unquoted value for "key: value" lines.
func extractYAMLField(content, key string) string {
for _, line := range strings.Split(content, "\n") {
if strings.HasPrefix(line, key+":") {
val := strings.TrimSpace(strings.TrimPrefix(line, key+":"))
return strings.Trim(val, `"`)
}
}
return ""
}
