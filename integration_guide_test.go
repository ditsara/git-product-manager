package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGuideFullOutput verifies that pm guide outputs all section headers.
func TestGuideFullOutput(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "guide")
	if err != nil {
		t.Fatalf("pm guide failed: %v\nOutput: %s", err, output)
	}

	expectedHeaders := []string{
		"# GPM Development Workflow",
		"# GPM Ticket Schema",
		"# GPM Commands Reference",
		"# GPM Key Principles",
	}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("pm guide missing expected header %q", header)
		}
	}
}

// TestGuideSectionWorkflow verifies pm guide workflow outputs only that section.
func TestGuideSectionWorkflow(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "guide", "workflow")
	if err != nil {
		t.Fatalf("pm guide workflow failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "# GPM Development Workflow") {
		t.Errorf("pm guide workflow missing expected header")
	}
	// Must NOT contain other section headers
	if strings.Contains(output, "# GPM Ticket Schema") {
		t.Errorf("pm guide workflow should not contain schema section")
	}
	if strings.Contains(output, "# GPM Commands Reference") {
		t.Errorf("pm guide workflow should not contain commands section")
	}
}

// TestGuideSectionSchema verifies pm guide schema outputs only that section.
func TestGuideSectionSchema(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "guide", "schema")
	if err != nil {
		t.Fatalf("pm guide schema failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "# GPM Ticket Schema") {
		t.Errorf("pm guide schema missing expected header")
	}
	if strings.Contains(output, "# GPM Development Workflow") {
		t.Errorf("pm guide schema should not contain workflow section")
	}
}

// TestGuidePrinciplesNoCommit verifies the no-commit principle is present.
func TestGuidePrinciplesNoCommit(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "guide", "principles")
	if err != nil {
		t.Fatalf("pm guide principles failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "commit") || !strings.Contains(output, "user") {
		t.Errorf("pm guide principles should mention not committing without user approval")
	}
}

// TestGuideInvalidSection verifies an unknown section returns an error.
func TestGuideInvalidSection(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "guide", "nonexistent")
	if err == nil {
		t.Fatalf("pm guide nonexistent should have failed, got output: %s", output)
	}

	if !strings.Contains(output, "nonexistent") {
		t.Errorf("error output should mention the invalid section name, got: %s", output)
	}
	if !strings.Contains(output, "workflow") {
		t.Errorf("error output should list valid section names, got: %s", output)
	}
}

// TestGuideFullIsPipeable verifies the full output is non-empty and ends with newline.
func TestGuideFullIsPipeable(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "guide")
	if err != nil {
		t.Fatalf("pm guide failed: %v\nOutput: %s", err, output)
	}
	if len(output) == 0 {
		t.Fatal("pm guide produced empty output")
	}
	if output[len(output)-1] != '\n' {
		t.Error("pm guide output should end with newline (pipeable)")
	}
}

// TestInitCreatesWorkflowGuide verifies pm init creates WORKFLOW_GUIDE.md stub.
func TestInitCreatesWorkflowGuide(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	guidePath := filepath.Join(workspace, ".pm", "config", "WORKFLOW_GUIDE.md")
	data, err := os.ReadFile(guidePath)
	if err != nil {
		t.Fatalf("WORKFLOW_GUIDE.md not created: %v", err)
	}

	content := string(data)
	// Must be a stub pointing to pm guide, not full content
	if !strings.Contains(content, "pm guide") {
		t.Errorf("WORKFLOW_GUIDE.md should reference 'pm guide', got: %s", content)
	}
	// Should list sections
	for _, section := range []string{"workflow", "schema", "commands", "principles"} {
		if !strings.Contains(content, section) {
			t.Errorf("WORKFLOW_GUIDE.md should mention section %q", section)
		}
	}
	// Should NOT be excessively long (it's a stub, not full content)
	if len(content) > 2000 {
		t.Errorf("WORKFLOW_GUIDE.md seems too long for a stub: %d chars", len(content))
	}
}
