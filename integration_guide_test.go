package main

import (
	"strings"
	"testing"
)

// TestGuideFullOutput verifies that pm ai guide outputs all section headers.
func TestGuideFullOutput(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "ai", "guide")
	if err != nil {
		t.Fatalf("pm ai guide failed: %v\nOutput: %s", err, output)
	}

	expectedHeaders := []string{
		"# GPM Development Workflow",
		"# GPM Ticket Schema",
		"# GPM Commands Reference",
		"# GPM Key Principles",
	}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("pm ai guide missing expected header %q", header)
		}
	}
}

// TestGuideSectionWorkflow verifies pm ai guide workflow outputs only that section.
func TestGuideSectionWorkflow(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "ai", "guide", "workflow")
	if err != nil {
		t.Fatalf("pm ai guide workflow failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "# GPM Development Workflow") {
		t.Errorf("pm ai guide workflow missing expected header")
	}
	// Must NOT contain other section headers
	if strings.Contains(output, "# GPM Ticket Schema") {
		t.Errorf("pm ai guide workflow should not contain schema section")
	}
	if strings.Contains(output, "# GPM Commands Reference") {
		t.Errorf("pm ai guide workflow should not contain commands section")
	}
}

// TestGuideSectionSchema verifies pm ai guide schema outputs only that section.
func TestGuideSectionSchema(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "ai", "guide", "schema")
	if err != nil {
		t.Fatalf("pm ai guide schema failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "# GPM Ticket Schema") {
		t.Errorf("pm ai guide schema missing expected header")
	}
	if strings.Contains(output, "# GPM Development Workflow") {
		t.Errorf("pm ai guide schema should not contain workflow section")
	}
}

// TestGuidePrinciplesNoCommit verifies the no-commit principle is present.
func TestGuidePrinciplesNoCommit(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "ai", "guide", "principles")
	if err != nil {
		t.Fatalf("pm ai guide principles failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "commit") || !strings.Contains(output, "user") {
		t.Errorf("pm ai guide principles should mention not committing without user approval")
	}
}

// TestGuideInvalidSection verifies an unknown section returns an error.
func TestGuideInvalidSection(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "GUIDE")

	output, err := runPM(t, pmBinary, workspace, "ai", "guide", "nonexistent")
	if err == nil {
		t.Fatalf("pm ai guide nonexistent should have failed, got output: %s", output)
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

	output, err := runPM(t, pmBinary, workspace, "ai", "guide")
	if err != nil {
		t.Fatalf("pm ai guide failed: %v\nOutput: %s", err, output)
	}
	if len(output) == 0 {
		t.Fatal("pm ai guide produced empty output")
	}
	if output[len(output)-1] != '\n' {
		t.Error("pm ai guide output should end with newline (pipeable)")
	}
}


