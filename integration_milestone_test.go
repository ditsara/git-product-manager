package main

import (
	"fmt"
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

// TestListMilestoneFilter verifies --milestone filtering in pm list.
func TestListMilestoneFilter(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	// Create a milestone
	_, err := runPM(t, pmBinary, workspace, "milestone", "create", "Sprint 1", "--due", "2026-06-01")
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}

	// Create two tickets
	_, err = runPM(t, pmBinary, workspace, "new", "Ticket A")
	if err != nil {
		t.Fatalf("failed to create Ticket A: %v", err)
	}
	_, err = runPM(t, pmBinary, workspace, "new", "Ticket B")
	if err != nil {
		t.Fatalf("failed to create Ticket B: %v", err)
	}

	// Assign first ticket to the milestone
	_, err = runPM(t, pmBinary, workspace, "edit", "TEST-1", "--field", "milestones=sprint-1")
	if err != nil {
		t.Fatalf("failed to assign milestone to TEST-1: %v", err)
	}

	t.Run("filter_shows_milestoned_ticket", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-1")
		if err != nil {
			t.Fatalf("pm list --milestone sprint-1 failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "TEST-1") {
			t.Errorf("expected output to contain 'TEST-1', got: %s", output)
		}
		if strings.Contains(output, "TEST-2") {
			t.Errorf("expected output NOT to contain 'TEST-2', got: %s", output)
		}
	})

	t.Run("combined_with_status_filter", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-1", "--status", "done")
		if err != nil {
			t.Fatalf("pm list --milestone sprint-1 --status done failed: %v\nOutput: %s", err, output)
		}
		// TEST-1 is not done, so it should not appear
		if strings.Contains(output, "TEST-1") {
			t.Errorf("expected output NOT to contain 'TEST-1' (not done), got: %s", output)
		}
	})

	t.Run("nonexistent_milestone_warns", func(t *testing.T) {
		output, _ := runPM(t, pmBinary, workspace, "list", "--milestone", "nonexistent")
		if !strings.Contains(output, "Warning") {
			t.Errorf("expected 'Warning' in output for nonexistent milestone, got: %s", output)
		}
	})
}

// TestMilestoneProgress tests progress display and close functionality.
func TestMilestoneProgress(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	// Create milestone "Sprint 1"
	_, err := runPM(t, pmBinary, workspace, "milestone", "create", "Sprint 1")
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}

	// Create 3 tickets and assign all to sprint-1
	for i := 0; i < 3; i++ {
		_, err = runPM(t, pmBinary, workspace, "new", "Sprint Ticket")
		if err != nil {
			t.Fatalf("failed to create ticket %d: %v", i+1, err)
		}
	}
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("TEST-%d", i)
		_, err = runPM(t, pmBinary, workspace, "edit", id, "--field", "milestones=sprint-1")
		if err != nil {
			t.Fatalf("failed to assign milestone to %s: %v", id, err)
		}
	}

	// Move 1 ticket to done
	_, err = runPM(t, pmBinary, workspace, "move", "TEST-1", "done")
	if err != nil {
		t.Fatalf("failed to move TEST-1 to done: %v", err)
	}

	t.Run("show_displays_progress", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "show", "sprint-1")
		if err != nil {
			t.Fatalf("pm milestone show sprint-1 failed: %v\nOutput: %s", err, output)
		}
		// Should contain progress info: 1/3 tickets
		if !strings.Contains(output, "1/3") && !strings.Contains(output, "33%") {
			t.Errorf("expected progress '1/3' or '33%%' in output, got: %s", output)
		}
	})

	t.Run("close_fails_with_incomplete", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "close", "sprint-1")
		if err == nil {
			t.Fatalf("expected error when closing with incomplete tickets, got success\nOutput: %s", output)
		}
		if !strings.Contains(output, "not done") {
			t.Errorf("expected 'not done' in output, got: %s", output)
		}
	})

	t.Run("close_force_succeeds", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "close", "sprint-1", "--force")
		if err != nil {
			t.Fatalf("pm milestone close --force failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "Closed") {
			t.Errorf("expected 'Closed' in output, got: %s", output)
		}
	})

	t.Run("show_state_is_closed", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "show", "sprint-1")
		if err != nil {
			t.Fatalf("pm milestone show sprint-1 failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "closed") {
			t.Errorf("expected state 'closed' in output, got: %s", output)
		}
	})
}

// TestMilestoneAddRemove tests pm milestone add and pm milestone remove.
func TestMilestoneAddRemove(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	_, err := runPM(t, pmBinary, workspace, "milestone", "create", "Sprint 1", "--id", "sprint-1")
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}
	_, err = runPM(t, pmBinary, workspace, "new", "Ticket A")
	if err != nil {
		t.Fatalf("failed to create Ticket A: %v", err)
	}
	_, err = runPM(t, pmBinary, workspace, "new", "Ticket B")
	if err != nil {
		t.Fatalf("failed to create Ticket B: %v", err)
	}

	t.Run("add_assigns_milestone", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "add", "sprint-1", "TEST-1")
		if err != nil {
			t.Fatalf("pm milestone add failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "TEST-1") {
			t.Errorf("expected output to mention TEST-1, got: %s", output)
		}
		// Verify milestone shows ticket
		listOut, err := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-1")
		if err != nil {
			t.Fatalf("pm list --milestone sprint-1 failed: %v\nOutput: %s", err, listOut)
		}
		if !strings.Contains(listOut, "TEST-1") {
			t.Errorf("expected TEST-1 in milestone list, got: %s", listOut)
		}
		if strings.Contains(listOut, "TEST-2") {
			t.Errorf("TEST-2 should not be in milestone list, got: %s", listOut)
		}
	})

	t.Run("add_is_idempotent", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "add", "sprint-1", "TEST-1")
		if err != nil {
			t.Fatalf("second pm milestone add failed: %v\nOutput: %s", err, output)
		}
		// No tickets modified (idempotent)
		if strings.Contains(output, "+ TEST-1") {
			t.Errorf("expected no modification on second add, got: %s", output)
		}
		if !strings.Contains(output, "0 ticket") {
			t.Errorf("expected '0 ticket(s)' for idempotent add, got: %s", output)
		}
	})

	t.Run("add_preserves_other_milestones", func(t *testing.T) {
		_, err := runPM(t, pmBinary, workspace, "milestone", "create", "Sprint 2", "--id", "sprint-2")
		if err != nil {
			t.Fatalf("failed to create sprint-2: %v", err)
		}
		_, err = runPM(t, pmBinary, workspace, "milestone", "add", "sprint-2", "TEST-1")
		if err != nil {
			t.Fatalf("failed to add sprint-2 to TEST-1: %v", err)
		}
		// TEST-1 should still be in sprint-1
		listOut, err := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-1")
		if err != nil {
			t.Fatalf("pm list --milestone sprint-1 failed: %v\nOutput: %s", err, listOut)
		}
		if !strings.Contains(listOut, "TEST-1") {
			t.Errorf("TEST-1 should still be in sprint-1 after adding sprint-2, got: %s", listOut)
		}
	})

	t.Run("remove_unassigns_milestone", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "remove", "sprint-1", "TEST-1")
		if err != nil {
			t.Fatalf("pm milestone remove failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "TEST-1") {
			t.Errorf("expected output to mention TEST-1, got: %s", output)
		}
		// Verify ticket is gone from milestone
		listOut, _ := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-1")
		if strings.Contains(listOut, "TEST-1") {
			t.Errorf("TEST-1 should not be in sprint-1 after remove, got: %s", listOut)
		}
		// But still in sprint-2
		listOut2, err := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-2")
		if err != nil {
			t.Fatalf("pm list --milestone sprint-2 failed: %v\nOutput: %s", err, listOut2)
		}
		if !strings.Contains(listOut2, "TEST-1") {
			t.Errorf("TEST-1 should still be in sprint-2, got: %s", listOut2)
		}
	})

	t.Run("remove_is_idempotent", func(t *testing.T) {
		// TEST-1 is already removed from sprint-1
		output, err := runPM(t, pmBinary, workspace, "milestone", "remove", "sprint-1", "TEST-1")
		if err != nil {
			t.Fatalf("second pm milestone remove failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "0 ticket") {
			t.Errorf("expected '0 ticket(s)' for idempotent remove, got: %s", output)
		}
	})

	t.Run("invalid_milestone_errors", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "add", "no-such-milestone", "TEST-1")
		if err == nil {
			t.Fatalf("expected error for invalid milestone, got success\nOutput: %s", output)
		}
		if !strings.Contains(output, "not found") {
			t.Errorf("expected 'not found' in error, got: %s", output)
		}
	})

	t.Run("invalid_ticket_errors", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "add", "sprint-1", "TEST-9999")
		if err == nil {
			t.Fatalf("expected error for invalid ticket, got success\nOutput: %s", output)
		}
		if !strings.Contains(output, "not found") {
			t.Errorf("expected 'not found' in error, got: %s", output)
		}
	})
}

// TestMilestoneAddRemoveCascade tests --cascade flag on add and remove.
func TestMilestoneAddRemoveCascade(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	_, err := runPM(t, pmBinary, workspace, "milestone", "create", "Sprint 1", "--id", "sprint-1")
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}

	// Create parent (epic) and two children
	_, err = runPM(t, pmBinary, workspace, "new", "Parent Epic", "--type", "epic")
	if err != nil {
		t.Fatalf("failed to create parent: %v", err)
	}
	_, err = runPM(t, pmBinary, workspace, "new", "Child A", "--parent", "TEST-1")
	if err != nil {
		t.Fatalf("failed to create child A: %v", err)
	}
	_, err = runPM(t, pmBinary, workspace, "new", "Child B", "--parent", "TEST-1")
	if err != nil {
		t.Fatalf("failed to create child B: %v", err)
	}
	// Grandchild under Child A
	_, err = runPM(t, pmBinary, workspace, "new", "Grandchild", "--parent", "TEST-2")
	if err != nil {
		t.Fatalf("failed to create grandchild: %v", err)
	}

	t.Run("cascade_add_includes_all_descendants", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "add", "sprint-1", "TEST-1", "--cascade")
		if err != nil {
			t.Fatalf("pm milestone add --cascade failed: %v\nOutput: %s", err, output)
		}
		// All 4 tickets should be added
		for _, id := range []string{"TEST-1", "TEST-2", "TEST-3", "TEST-4"} {
			if !strings.Contains(output, id) {
				t.Errorf("expected %s in cascade add output, got: %s", id, output)
			}
		}
		if !strings.Contains(output, "4 ticket") {
			t.Errorf("expected '4 ticket(s)' in output, got: %s", output)
		}
		// Verify via list
		listOut, err := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-1")
		if err != nil {
			t.Fatalf("pm list --milestone sprint-1 failed: %v\nOutput: %s", err, listOut)
		}
		for _, id := range []string{"TEST-1", "TEST-2", "TEST-3", "TEST-4"} {
			if !strings.Contains(listOut, id) {
				t.Errorf("expected %s in milestone list after cascade add, got: %s", id, listOut)
			}
		}
	})

	t.Run("cascade_remove_removes_all_descendants", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "milestone", "remove", "sprint-1", "TEST-1", "--cascade")
		if err != nil {
			t.Fatalf("pm milestone remove --cascade failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "4 ticket") {
			t.Errorf("expected '4 ticket(s)' in remove cascade output, got: %s", output)
		}
		// Verify milestone is empty
		listOut, _ := runPM(t, pmBinary, workspace, "list", "--milestone", "sprint-1")
		for _, id := range []string{"TEST-1", "TEST-2", "TEST-3", "TEST-4"} {
			if strings.Contains(listOut, id) {
				t.Errorf("%s should not be in sprint-1 after cascade remove, got: %s", id, listOut)
			}
		}
	})

	t.Run("cascade_add_non_root_ticket", func(t *testing.T) {
		// Add from Child A (TEST-2) — should include TEST-2 and TEST-4 (grandchild), not TEST-3
		output, err := runPM(t, pmBinary, workspace, "milestone", "add", "sprint-1", "TEST-2", "--cascade")
		if err != nil {
			t.Fatalf("pm milestone add --cascade TEST-2 failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "TEST-2") {
			t.Errorf("expected TEST-2 in output, got: %s", output)
		}
		if !strings.Contains(output, "TEST-4") {
			t.Errorf("expected TEST-4 (grandchild) in output, got: %s", output)
		}
		if strings.Contains(output, "TEST-3") {
			t.Errorf("TEST-3 should not be included (different branch), got: %s", output)
		}
	})
}

// TestMilestoneCloseAllDone tests that close succeeds without --force when all tickets are done.
func TestMilestoneCloseAllDone(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "TEST")

	_, err := runPM(t, pmBinary, workspace, "milestone", "create", "Done Sprint", "--id", "done-sprint")
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}

	_, err = runPM(t, pmBinary, workspace, "new", "Only Ticket")
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}

	_, err = runPM(t, pmBinary, workspace, "edit", "TEST-1", "--field", "milestones=done-sprint")
	if err != nil {
		t.Fatalf("failed to assign milestone: %v", err)
	}

	_, err = runPM(t, pmBinary, workspace, "move", "TEST-1", "done")
	if err != nil {
		t.Fatalf("failed to move ticket to done: %v", err)
	}

	output, err := runPM(t, pmBinary, workspace, "milestone", "close", "done-sprint")
	if err != nil {
		t.Fatalf("close should succeed when all tickets done: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Closed") {
		t.Errorf("expected 'Closed' in output, got: %s", output)
	}
}
