package main

import (
	"strings"
	"testing"
)

// TestTreeSimpleHierarchy tests tree display with a simple parent-child structure
func TestTreeSimpleHierarchy(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "TREE")

	// Create parent ticket
	runPM(t, pmBinary, workspace, "new", "--type", "epic", "Parent Epic")
	// Create child tickets
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-1", "Child Task 1")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-1", "Child Task 2")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-1", "Child Task 3")

	// Run tree command
	output, err := runPM(t, pmBinary, workspace, "tree", "TREE-1")
	if err != nil {
		t.Fatalf("pm tree failed: %v\nOutput: %s", err, output)
	}

	// Verify root ticket is displayed
	if !strings.Contains(output, "TREE-1: Parent Epic") {
		t.Errorf("Output missing root ticket: %s", output)
	}

	// Verify all children are displayed
	if !strings.Contains(output, "TREE-2: Child Task 1") {
		t.Errorf("Output missing child 1: %s", output)
	}
	if !strings.Contains(output, "TREE-3: Child Task 2") {
		t.Errorf("Output missing child 2: %s", output)
	}
	if !strings.Contains(output, "TREE-4: Child Task 3") {
		t.Errorf("Output missing child 3: %s", output)
	}

	// Verify box-drawing characters are present
	if !strings.Contains(output, "├──") && !strings.Contains(output, "└──") {
		t.Errorf("Output missing box-drawing characters: %s", output)
	}
}

// TestTreeMultiLevelHierarchy tests tree with multiple levels of nesting
func TestTreeMultiLevelHierarchy(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "TREE")

	// Create nested structure:
	// TREE-1 (Epic)
	//   ├── TREE-2 (Story)
	//   │   ├── TREE-4 (Task)
	//   │   └── TREE-5 (Task)
	//   └── TREE-3 (Story)
	//       └── TREE-6 (Task)

	runPM(t, pmBinary, workspace, "new", "--type", "epic", "Root Epic")
	runPM(t, pmBinary, workspace, "new", "--type", "story", "--parent", "TREE-1", "Story 1")
	runPM(t, pmBinary, workspace, "new", "--type", "story", "--parent", "TREE-1", "Story 2")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-2", "Task 1-1")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-2", "Task 1-2")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-3", "Task 2-1")

	// Run tree from root
	output, err := runPM(t, pmBinary, workspace, "tree", "TREE-1")
	if err != nil {
		t.Fatalf("pm tree failed: %v\nOutput: %s", err, output)
	}

	// Verify all levels are present
	if !strings.Contains(output, "TREE-1: Root Epic") {
		t.Errorf("Output missing root: %s", output)
	}
	if !strings.Contains(output, "TREE-2: Story 1") {
		t.Errorf("Output missing story 1: %s", output)
	}
	if !strings.Contains(output, "TREE-3: Story 2") {
		t.Errorf("Output missing story 2: %s", output)
	}
	if !strings.Contains(output, "TREE-4: Task 1-1") {
		t.Errorf("Output missing task 1-1: %s", output)
	}
	if !strings.Contains(output, "TREE-5: Task 1-2") {
		t.Errorf("Output missing task 1-2: %s", output)
	}
	if !strings.Contains(output, "TREE-6: Task 2-1") {
		t.Errorf("Output missing task 2-1: %s", output)
	}

	// Verify continuation lines (pipes and spaces)
	if !strings.Contains(output, "│") {
		t.Errorf("Output missing continuation pipes: %s", output)
	}
}

// TestTreeNoChildren tests tree with ticket that has no children
func TestTreeNoChildren(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "TREE")

	// Create a single ticket with no children
	runPM(t, pmBinary, workspace, "new", "--type", "task", "Standalone Task")

	// Run tree command
	output, err := runPM(t, pmBinary, workspace, "tree", "TREE-1")
	if err != nil {
		t.Fatalf("pm tree failed: %v\nOutput: %s", err, output)
	}

	// Verify ticket is shown
	if !strings.Contains(output, "TREE-1: Standalone Task") {
		t.Errorf("Output missing ticket: %s", output)
	}
}

// TestTreeDepthLimit tests --depth flag limiting output
func TestTreeDepthLimit(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "TREE")

	// Create nested structure: 3 levels
	// TREE-1
	//   └── TREE-2
	//       └── TREE-3

	runPM(t, pmBinary, workspace, "new", "--type", "epic", "Level 1")
	runPM(t, pmBinary, workspace, "new", "--type", "story", "--parent", "TREE-1", "Level 2")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-2", "Level 3")

	// Run tree with --depth 1 (should only show root)
	output, err := runPM(t, pmBinary, workspace, "tree", "TREE-1", "--depth", "1")
	if err != nil {
		t.Fatalf("pm tree --depth 1 failed: %v\nOutput: %s", err, output)
	}

	// Should show root
	if !strings.Contains(output, "TREE-1: Level 1") {
		t.Errorf("Output missing root: %s", output)
	}

	// With --depth 1, should not show children (depth 2 and beyond)
	if strings.Contains(output, "TREE-2: Level 2") || strings.Contains(output, "TREE-3: Level 3") {
		t.Errorf("Output should not contain children with --depth 1: %s", output)
	}

	// Run tree with --depth 2 (should show root and direct children)
	output, err = runPM(t, pmBinary, workspace, "tree", "TREE-1", "--depth", "2")
	if err != nil {
		t.Fatalf("pm tree --depth 2 failed: %v\nOutput: %s", err, output)
	}

	// Should show root and level 2
	if !strings.Contains(output, "TREE-1: Level 1") {
		t.Errorf("Output missing root: %s", output)
	}
	if !strings.Contains(output, "TREE-2: Level 2") {
		t.Errorf("Output missing level 2: %s", output)
	}

	// Should not show level 3
	if strings.Contains(output, "TREE-3: Level 3") {
		t.Errorf("Output should not contain level 3 with --depth 2: %s", output)
	}
}

// TestTreeNonexistentTicket tests error handling for non-existent ticket
func TestTreeNonexistentTicket(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "TREE")

	// Try to display tree for non-existent ticket
	_, err := runPM(t, pmBinary, workspace, "tree", "TREE-999")

	// Should error
	if err == nil {
		t.Errorf("pm tree should fail for non-existent ticket")
	}
}

// TestTreeCaseMismatch tests case-insensitive ticket ID lookup
func TestTreeCaseMismatch(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "TREE")

	// Create parent and child
	runPM(t, pmBinary, workspace, "new", "--type", "epic", "Parent")
	runPM(t, pmBinary, workspace, "new", "--type", "task", "--parent", "TREE-1", "Child")

	// Run tree with lowercase ticket ID
	output, err := runPM(t, pmBinary, workspace, "tree", "tree-1")
	if err != nil {
		t.Fatalf("pm tree with lowercase ID failed: %v\nOutput: %s", err, output)
	}

	// Should still work and display tree correctly
	if !strings.Contains(output, "TREE-1: Parent") {
		t.Errorf("Output missing root with lowercase lookup: %s", output)
	}
	if !strings.Contains(output, "TREE-2: Child") {
		t.Errorf("Output missing child with lowercase lookup: %s", output)
	}
}

// TestTreeLongTitleTruncation tests that long titles are truncated
func TestTreeLongTitleTruncation(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "TREE")

	// Create ticket with very long title (more than 60 chars)
	longTitle := "This is a very long ticket title that exceeds the maximum truncation limit of sixty characters"
	runPM(t, pmBinary, workspace, "new", "--type", "task", longTitle)

	// Run tree command
	output, err := runPM(t, pmBinary, workspace, "tree", "TREE-1")
	if err != nil {
		t.Fatalf("pm tree failed: %v\nOutput: %s", err, output)
	}

	// Should have truncated the title (not contain full long title)
	if strings.Contains(output, longTitle) {
		t.Errorf("Output should not contain untruncated long title: %s", output)
	}

	// But should show the start of the title and ellipsis
	if !strings.Contains(output, "…") {
		t.Errorf("Output should contain ellipsis for truncated title: %s", output)
	}
}
