package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkflow(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "workflow.yaml")

	validWorkflow := `states:
  - backlog
  - todo
  - in-progress
  - done

initial_state: backlog

state_groups:
  active: [todo, in-progress]
  completed: [done]
  incomplete: [backlog, todo, in-progress]
`

	if err := os.WriteFile(workflowPath, []byte(validWorkflow), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	workflow, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() unexpected error = %v", err)
	}

	if workflow.InitialState != "backlog" {
		t.Errorf("LoadWorkflow() InitialState = %v, want 'backlog'", workflow.InitialState)
	}

	expectedStates := []string{"backlog", "todo", "in-progress", "done"}
	if len(workflow.States) != len(expectedStates) {
		t.Errorf("LoadWorkflow() States count = %v, want %v", len(workflow.States), len(expectedStates))
	}

	for i, state := range expectedStates {
		if workflow.States[i] != state {
			t.Errorf("LoadWorkflow() States[%d] = %v, want %v", i, workflow.States[i], state)
		}
	}

	// Test state groups
	if len(workflow.StateGroups["active"]) != 2 {
		t.Errorf("LoadWorkflow() StateGroups['active'] count = %v, want 2", len(workflow.StateGroups["active"]))
	}
	if len(workflow.StateGroups["completed"]) != 1 {
		t.Errorf("LoadWorkflow() StateGroups['completed'] count = %v, want 1", len(workflow.StateGroups["completed"]))
	}
}

func TestLoadWorkflowInvalidFile(t *testing.T) {
	_, err := LoadWorkflow("/nonexistent/path/workflow.yaml")
	if err == nil {
		t.Error("LoadWorkflow() expected error for nonexistent file, got nil")
	}
}

func TestLoadWorkflowInvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "workflow.yaml")

	invalidYAML := `this is not: [[[valid yaml`
	if err := os.WriteFile(workflowPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := LoadWorkflow(workflowPath)
	if err == nil {
		t.Error("LoadWorkflow() expected error for invalid YAML, got nil")
	}
}

func TestIsValidState(t *testing.T) {
	workflow := &Workflow{
		States:       []string{"backlog", "todo", "in-progress", "done"},
		InitialState: "backlog",
	}

	tests := []struct {
		name  string
		state string
		valid bool
	}{
		{"valid backlog", "backlog", true},
		{"valid todo", "todo", true},
		{"valid in-progress", "in-progress", true},
		{"valid done", "done", true},
		{"invalid completed", "completed", false},
		{"invalid empty", "", false},
		{"invalid random", "random-state", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := workflow.IsValidState(tt.state)
			if result != tt.valid {
				t.Errorf("IsValidState(%q) = %v, want %v", tt.state, result, tt.valid)
			}
		})
	}
}

func TestIsCompleted(t *testing.T) {
	workflow := &Workflow{
		States:       []string{"backlog", "todo", "in-progress", "done", "canceled"},
		InitialState: "backlog",
		StateGroups: map[string][]string{
			"active":     {"todo", "in-progress"},
			"completed":  {"done", "canceled"},
			"incomplete": {"backlog", "todo", "in-progress"},
		},
	}

	tests := []struct {
		name      string
		status    string
		completed bool
	}{
		{"done is completed", "done", true},
		{"canceled is completed", "canceled", true},
		{"backlog is not completed", "backlog", false},
		{"todo is not completed", "todo", false},
		{"in-progress is not completed", "in-progress", false},
		{"unknown state is not completed", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := workflow.IsCompleted(tt.status)
			if result != tt.completed {
				t.Errorf("IsCompleted(%q) = %v, want %v", tt.status, result, tt.completed)
			}
		})
	}
}

func TestGetCompletedStates(t *testing.T) {
	t.Run("with completed group", func(t *testing.T) {
		workflow := &Workflow{
			States:       []string{"backlog", "todo", "done"},
			InitialState: "backlog",
			StateGroups: map[string][]string{
				"completed": {"done"},
			},
		}

		states := workflow.GetCompletedStates()
		if len(states) != 1 || states[0] != "done" {
			t.Errorf("GetCompletedStates() = %v, want [done]", states)
		}
	})

	t.Run("without completed group", func(t *testing.T) {
		workflow := &Workflow{
			States:       []string{"backlog", "todo", "done"},
			InitialState: "backlog",
			StateGroups:  map[string][]string{},
		}

		states := workflow.GetCompletedStates()
		if states != nil {
			t.Errorf("GetCompletedStates() = %v, want nil", states)
		}
	})
}

func TestGetStateGroup(t *testing.T) {
	workflow := &Workflow{
		States:       []string{"backlog", "todo", "in-progress", "done"},
		InitialState: "backlog",
		StateGroups: map[string][]string{
			"active":     {"todo", "in-progress"},
			"completed":  {"done"},
			"incomplete": {"backlog", "todo", "in-progress"},
		},
	}

	tests := []struct {
		name      string
		groupName string
		expected  []string
	}{
		{"active group", "active", []string{"todo", "in-progress"}},
		{"completed group", "completed", []string{"done"}},
		{"incomplete group", "incomplete", []string{"backlog", "todo", "in-progress"}},
		{"nonexistent group", "blocked", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := workflow.GetStateGroup(tt.groupName)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetStateGroup(%q) = %v, want nil", tt.groupName, result)
				}
			} else {
				if len(result) != len(tt.expected) {
					t.Errorf("GetStateGroup(%q) length = %d, want %d", tt.groupName, len(result), len(tt.expected))
					return
				}
				for i, state := range tt.expected {
					if result[i] != state {
						t.Errorf("GetStateGroup(%q)[%d] = %q, want %q", tt.groupName, i, result[i], state)
					}
				}
			}
		})
	}
}
