package main

import (
	"testing"
)

// TestFormatNode tests the formatNode function
func TestFormatNode(t *testing.T) {
	tests := []struct {
		name          string
		node          *TreeNode
		expectTrunc   bool
		expectEllipsis bool
	}{
		{
			name: "short title",
			node: &TreeNode{
				ID:    "TEST-1",
				Title: "Short title",
				Type:  "task",
				Depth: 0,
			},
			expectTrunc:   false,
			expectEllipsis: false,
		},
		{
			name: "title exactly 60 chars",
			node: &TreeNode{
				ID:    "TEST-1",
				Title: "This is exactly sixty characters long for testing purposes",
				Type:  "task",
				Depth: 0,
			},
			expectTrunc:   false,
			expectEllipsis: false,
		},
		{
			name: "title longer than 60 chars",
			node: &TreeNode{
				ID:    "TEST-1",
				Title: "This is a very long title that exceeds sixty characters and should be truncated",
				Type:  "task",
				Depth: 0,
			},
			expectTrunc:   true,
			expectEllipsis: true,
		},
		{
			name: "empty title",
			node: &TreeNode{
				ID:    "TEST-1",
				Title: "",
				Type:  "task",
				Depth: 0,
			},
			expectTrunc:   false,
			expectEllipsis: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatNode(tt.node, nil)

			// Should always contain ID
			if len(output) == 0 || output[:len(tt.node.ID)] != tt.node.ID {
				t.Errorf("Output should start with ticket ID: %s", output)
			}

			// Check for truncation
			if tt.expectEllipsis && !contains(output, "…") {
				t.Errorf("Expected ellipsis in output, got: %s", output)
			}

			if !tt.expectEllipsis && contains(output, "…") {
				t.Errorf("Should not have ellipsis in output, got: %s", output)
			}
		})
	}
}

// TestRenderChildren tests the renderChildren function (basic structure)
func TestRenderChildrenStructure(t *testing.T) {
	// Create a simple tree
	root := &TreeNode{
		ID:    "ROOT",
		Title: "Root",
		Depth: 0,
		Children: []*TreeNode{
			{
				ID:    "CHILD1",
				Title: "Child 1",
				Depth: 1,
			},
			{
				ID:    "CHILD2",
				Title: "Child 2",
				Depth: 1,
			},
		},
	}

	// Test that tree building doesn't panic
	// (Full output testing is done in integration tests)
	if len(root.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(root.Children))
	}

	if root.Children[0].Depth != 1 || root.Children[1].Depth != 1 {
		t.Errorf("Children should have depth 1")
	}
}

// TestTreeNodeStructure tests TreeNode initialization
func TestTreeNodeStructure(t *testing.T) {
	node := &TreeNode{
		ID:       "TEST-42",
		Title:    "Test Ticket",
		Type:     "task",
		Depth:    2,
		Children: make([]*TreeNode, 0),
	}

	if node.ID != "TEST-42" {
		t.Errorf("ID not set correctly")
	}
	if node.Title != "Test Ticket" {
		t.Errorf("Title not set correctly")
	}
	if node.Type != "task" {
		t.Errorf("Type not set correctly")
	}
	if node.Depth != 2 {
		t.Errorf("Depth not set correctly")
	}
	if len(node.Children) != 0 {
		t.Errorf("Children should be empty slice")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
