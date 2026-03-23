package main

import (
	"strings"
	"testing"
)

// TestMilestoneCreateListShow tests the full milestone create → list → show workflow.
func TestMilestoneCreateListShow(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	t.Run("create_milestone", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "create", "Version 1.0", "--due", "2026-12-31")
		if err != nil {
			t.Fatalf("pm milestone create failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "version-1-0") {
			t.Errorf("expected output to contain 'version-1-0', got: %s", output)
		}
		if !strings.Contains(output, "Created milestone") {
			t.Errorf("expected output to contain 'Created milestone', got: %s", output)
		}
	})

	t.Run("list_shows_milestone", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "list")
		if err != nil {
			t.Fatalf("pm milestone list failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "version-1-0") {
			t.Errorf("expected output to contain 'version-1-0', got: %s", output)
		}
		if !strings.Contains(output, "Version 1.0") {
			t.Errorf("expected output to contain 'Version 1.0', got: %s", output)
		}
		if !strings.Contains(output, "active") {
			t.Errorf("expected output to contain 'active', got: %s", output)
		}
	})

	t.Run("show_milestone", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "show", "version-1-0")
		if err != nil {
			t.Fatalf("pm milestone show failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "version-1-0") {
			t.Errorf("expected output to contain 'version-1-0', got: %s", output)
		}
		if !strings.Contains(output, "Version 1.0") {
			t.Errorf("expected output to contain 'Version 1.0', got: %s", output)
		}
		if !strings.Contains(output, "active") {
			t.Errorf("expected output to contain state 'active', got: %s", output)
		}
		if !strings.Contains(output, "Dec 31, 2026") {
			t.Errorf("expected output to contain due date 'Dec 31, 2026', got: %s", output)
		}
		if !strings.Contains(output, "Tickets:") {
			t.Errorf("expected output to contain 'Tickets:', got: %s", output)
		}
	})

	t.Run("duplicate_id_error", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "create", "Version 1.0")
		if err == nil {
			t.Fatalf("expected error for duplicate milestone ID, got success\nOutput: %s", output)
		}
		if !strings.Contains(output, "already exists") {
			t.Errorf("expected 'already exists' in error output, got: %s", output)
		}
	})

	t.Run("create_with_custom_id", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "create", "Version 1.0", "--id", "v1-custom")
		if err != nil {
			t.Fatalf("pm milestone create --id v1-custom failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "v1-custom") {
			t.Errorf("expected output to contain 'v1-custom', got: %s", output)
		}
	})
}

// TestMilestoneListEmpty verifies graceful output when no milestones exist.
func TestMilestoneListEmpty(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	output, err := runPM(t, pmBinary, workspace, "milestone", "list")
	if err != nil {
		t.Fatalf("pm milestone list failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "No milestones found") {
		t.Errorf("expected 'No milestones found', got: %s", output)
	}
}

// TestMilestoneStateFilter verifies --state filtering.
func TestMilestoneStateFilter(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	// Create two milestones
	_, err := runPM(t, pmBinary, workspace, "milestone", "create", "Active MS", "--id", "active-ms")
	if err != nil {
		t.Fatalf("failed to create active milestone: %v", err)
	}

	// Filter by active
	output, err := runPM(t, pmBinary, workspace, "milestone", "list", "--state", "active")
	if err != nil {
		t.Fatalf("pm milestone list --state active failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "active-ms") {
		t.Errorf("expected 'active-ms' in active list, got: %s", output)
	}

	// Filter by closed — should be empty
	output, err = runPM(t, pmBinary, workspace, "milestone", "list", "--state", "closed")
	if err != nil {
		t.Fatalf("pm milestone list --state closed failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "No milestones found") {
		t.Errorf("expected 'No milestones found' for closed filter, got: %s", output)
	}
}

// TestMilestoneShowNotFound verifies error output when milestone does not exist.
func TestMilestoneShowNotFound(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	output, err := runPM(t, pmBinary, workspace, "milestone", "show", "nonexistent")
	if err == nil {
		t.Fatalf("expected error for nonexistent milestone, got success\nOutput: %s", output)
	}
	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found' in error output, got: %s", output)
	}
}
