package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIInitGeneratesAgentsMD(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)

	// pm init should create .pm/AGENTS.md
	initWorkspace(t, pmBinary, workspace, "TST")

	agentsPath := filepath.Join(workspace, ".pm", "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("AGENTS.md not created by pm init: %v", err)
	}
	content := string(data)
	for _, want := range []string{"pm --help", "pm ai guide", "pm list", "pm show"} {
		if !strings.Contains(content, want) {
			t.Errorf("AGENTS.md missing expected content %q", want)
		}
	}
}

func TestAIInitIdempotentAgentsMD(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)

	// Run pm init twice
	initWorkspace(t, pmBinary, workspace, "TST")
	// Second init should fail (already exists) — AGENTS.md should not be replaced
	agentsPath := filepath.Join(workspace, ".pm", "AGENTS.md")
	first, _ := os.ReadFile(agentsPath)

	runPM(t, pmBinary, workspace, "init", ".", "--prefix", "TST") //nolint
	second, _ := os.ReadFile(agentsPath)

	if string(first) != string(second) {
		t.Errorf("AGENTS.md content changed on second pm init")
	}
}

func TestAIInitCommandAppendsSingle(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TST")

	out, err := runPM(t, pmBinary, workspace, "ai", "init", "--for", "claude")
	if err != nil {
		t.Fatalf("pm ai init --for claude failed: %v\nOutput: %s", err, out)
	}

	claudePath := filepath.Join(workspace, "CLAUDE.md")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	if !strings.Contains(string(data), ".pm/AGENTS.md") {
		t.Errorf("CLAUDE.md does not contain pointer to .pm/AGENTS.md")
	}
	// Only CLAUDE.md should have been created
	if _, err := os.Stat(filepath.Join(workspace, ".github", "copilot-instructions.md")); err == nil {
		t.Errorf("copilot-instructions.md should not be created when --for claude")
	}
}

func TestAIInitCommandAppendsAll(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TST")

	out, err := runPM(t, pmBinary, workspace, "ai", "init")
	if err != nil {
		t.Fatalf("pm ai init failed: %v\nOutput: %s", err, out)
	}

	wants := []string{
		"CLAUDE.md",
		filepath.Join(".github", "copilot-instructions.md"),
		filepath.Join(".cursor", "rules", "gpm.mdc"),
		"CONVENTIONS.md",
	}
	for _, rel := range wants {
		path := filepath.Join(workspace, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s not created: %v", rel, err)
			continue
		}
		if !strings.Contains(string(data), ".pm/AGENTS.md") {
			t.Errorf("%s does not contain pointer to .pm/AGENTS.md", rel)
		}
	}
}

func TestAIInitCommandIdempotent(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TST")

	// Run pm ai init twice for claude
	runPM(t, pmBinary, workspace, "ai", "init", "--for", "claude") //nolint
	runPM(t, pmBinary, workspace, "ai", "init", "--for", "claude") //nolint

	claudePath := filepath.Join(workspace, "CLAUDE.md")
	data, _ := os.ReadFile(claudePath)
	content := string(data)

	// Pointer block should only appear once
	count := strings.Count(content, "See .pm/AGENTS.md")
	if count != 1 {
		t.Errorf("Expected pointer to appear once, got %d occurrences", count)
	}
}

func TestAIInitCommandAppendsToExisting(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TST")

	// Pre-create CLAUDE.md with user content
	claudePath := filepath.Join(workspace, "CLAUDE.md")
	userContent := "# My existing CLAUDE.md\nSome user content here.\n"
	if err := os.WriteFile(claudePath, []byte(userContent), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := runPM(t, pmBinary, workspace, "ai", "init", "--for", "claude")
	if err != nil {
		t.Fatalf("pm ai init failed: %v\nOutput: %s", err, out)
	}

	data, _ := os.ReadFile(claudePath)
	content := string(data)

	if !strings.Contains(content, userContent) {
		t.Error("Existing content was removed")
	}
	if !strings.Contains(content, ".pm/AGENTS.md") {
		t.Error("Pointer was not appended")
	}
	if !strings.HasPrefix(content, userContent) {
		t.Error("Existing content was not preserved at start of file")
	}
}

func TestAIInitCommandUnknownTool(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TST")

	out, err := runPM(t, pmBinary, workspace, "ai", "init", "--for", "unknown-tool")
	if err == nil {
		t.Fatalf("Expected error for unknown tool, got success\nOutput: %s", out)
	}
	if !strings.Contains(out, "unknown tool") {
		t.Errorf("Expected 'unknown tool' in error output, got: %s", out)
	}
}

func TestInitIdempotent(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)

	// First init
	initWorkspace(t, pmBinary, workspace, "TST")

	// Capture state of all expected files
	files := []string{
		filepath.Join(".pm", "config", "project.yaml"),
		filepath.Join(".pm", "config", "workflow.yaml"),
		filepath.Join(".pm", "config", "labels.yaml"),
		filepath.Join(".pm", "config", "WORKFLOW_GUIDE.md"),
		filepath.Join(".pm", "AGENTS.md"),
	}
	snapshots := make(map[string]string)
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(workspace, rel))
		if err != nil {
			t.Fatalf("Missing expected file %s after first init: %v", rel, err)
		}
		snapshots[rel] = string(data)
	}

	// Second init — should succeed (not exit 1) and not overwrite existing files
	out, err := runPM(t, pmBinary, workspace, "init", ".", "--prefix", "TST")
	if err != nil {
		t.Fatalf("Second pm init failed: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "already initialised") {
		t.Errorf("Expected re-init message, got: %s", out)
	}

	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(workspace, rel))
		if err != nil {
			t.Fatalf("File %s disappeared after second init: %v", rel, err)
		}
		if string(data) != snapshots[rel] {
			t.Errorf("File %s was modified by second init", rel)
		}
	}
}

func TestInitRegeneratesMissingFile(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)

	initWorkspace(t, pmBinary, workspace, "TST")

	// Delete AGENTS.md
	agentsPath := filepath.Join(workspace, ".pm", "AGENTS.md")
	if err := os.Remove(agentsPath); err != nil {
		t.Fatal(err)
	}

	// Re-run init — should regenerate the missing file
	out, err := runPM(t, pmBinary, workspace, "init", ".", "--prefix", "TST")
	if err != nil {
		t.Fatalf("Re-init after delete failed: %v\nOutput: %s", err, out)
	}
	if _, err := os.Stat(agentsPath); err != nil {
		t.Errorf("AGENTS.md was not regenerated by re-init: %v", err)
	}
}
