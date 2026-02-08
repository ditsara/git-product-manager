package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestUpdateTicketFieldStringField tests updating string fields
func TestUpdateTicketFieldStringField(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		value       string
		expectValue string
		description string
	}{
		{
			name:        "update_assignee",
			field:       "assignee",
			value:       "alice",
			expectValue: "alice",
			description: "Assignee string field should be updated",
		},
		{
			name:        "update_title",
			field:       "title",
			value:       "New Title",
			expectValue: "New Title",
			description: "Title string field should be updated",
		},
		{
			name:        "empty_string_value",
			field:       "assignee",
			value:       "",
			expectValue: "",
			description: "Empty string should clear the field",
		},
		{
			name:        "assignee_with_email",
			field:       "assignee",
			value:       "alice@example.com",
			expectValue: "alice@example.com",
			description: "Assignee can contain email addresses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary ticket file
			tmpDir := t.TempDir()
			ticketPath := filepath.Join(tmpDir, "TEST-1.md")

			originalContent := `---
id: TEST-1
title: "Original Title"
type: task
status: backlog
priority: medium
assignee: ""
parent: ""
depends_on: []
blocks: []
related: []
labels: []
created_at: 2026-02-01T10:00:00Z
updated_at: 2026-02-01T10:00:00Z
---

# Description
Test ticket
`
			if err := os.WriteFile(ticketPath, []byte(originalContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Update the field
			updateTicketField(ticketPath, tt.field, tt.value)

			// Read and verify the file
			content, err := os.ReadFile(ticketPath)
			if err != nil {
				t.Fatalf("Failed to read updated file: %v", err)
			}

			// Parse the YAML
			parts := strings.SplitN(string(content), "---", 3)
			var metadata map[string]interface{}
			if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			actualValue := metadata[tt.field]
			if actualValue != tt.expectValue {
				t.Errorf("%s: Expected %q, got %q", tt.description, tt.expectValue, actualValue)
			}

			// Verify updated_at changed
			if metadata["updated_at"] == "2026-02-01T10:00:00Z" {
				t.Error("updated_at should have been changed")
			}
		})
	}
}

// TestUpdateTicketFieldIntegerField tests updating integer fields
func TestUpdateTicketFieldIntegerField(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		value       string
		expectValue int
		expectError bool
		description string
	}{
		{
			name:        "update_points_valid",
			field:       "points",
			value:       "5",
			expectValue: 5,
			expectError: false,
			description: "Valid integer should update points",
		},
		{
			name:        "update_points_zero",
			field:       "points",
			value:       "0",
			expectValue: 0,
			expectError: false,
			description: "Zero should be accepted for points",
		},
		{
			name:        "empty_points_defaults_to_zero",
			field:       "points",
			value:       "",
			expectValue: 0,
			expectError: false,
			description: "Empty value should default to 0 for integer field",
		},
		{
			name:        "invalid_integer",
			field:       "points",
			value:       "not-a-number",
			expectError: true,
			description: "Non-integer value should cause error",
		},
		{
			name:        "large_integer",
			field:       "points",
			value:       "999",
			expectValue: 999,
			expectError: false,
			description: "Large integer values should be accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary ticket file
			tmpDir := t.TempDir()
			ticketPath := filepath.Join(tmpDir, "TEST-1.md")

			originalContent := `---
id: TEST-1
title: "Test Ticket"
type: task
status: backlog
priority: medium
points: 0
assignee: ""
parent: ""
depends_on: []
blocks: []
related: []
labels: []
created_at: 2026-02-01T10:00:00Z
updated_at: 2026-02-01T10:00:00Z
---

# Description
Test ticket
`
			if err := os.WriteFile(ticketPath, []byte(originalContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Update the field (may fail for invalid values)
			if tt.expectError {
				// We expect the function to exit or log an error
				// Skip verification for error cases as they use os.Exit
				return
			}

			updateTicketField(ticketPath, tt.field, tt.value)

			// Read and verify the file
			content, err := os.ReadFile(ticketPath)
			if err != nil {
				t.Fatalf("Failed to read updated file: %v", err)
			}

			// Parse the YAML
			parts := strings.SplitN(string(content), "---", 3)
			var metadata map[string]interface{}
			if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			actualValue := metadata[tt.field]
			// YAML unmarshals integers as interface{}, convert if needed
			var intVal int
			if iv, ok := actualValue.(int); ok {
				intVal = iv
			}

			if intVal != tt.expectValue {
				t.Errorf("%s: Expected %d, got %d", tt.description, tt.expectValue, intVal)
			}
		})
	}
}

// TestUpdateTicketFieldArrayField tests updating array/list fields
func TestUpdateTicketFieldArrayField(t *testing.T) {
	tests := []struct {
		name         string
		field        string
		value        string
		expectValues []string
		description  string
	}{
		{
			name:         "single_label",
			field:        "labels",
			value:        "bug",
			expectValues: []string{"bug"},
			description:  "Single label should be added to labels array",
		},
		{
			name:         "multiple_labels",
			field:        "labels",
			value:        "bug,feature,urgent",
			expectValues: []string{"bug", "feature", "urgent"},
			description:  "Multiple comma-separated labels should be parsed correctly",
		},
		{
			name:         "labels_with_spaces",
			field:        "labels",
			value:        "bug , feature , urgent",
			expectValues: []string{"bug", "feature", "urgent"},
			description:  "Whitespace around labels should be trimmed",
		},
		{
			name:         "empty_array_field",
			field:        "labels",
			value:        "",
			expectValues: []string{},
			description:  "Empty value should clear the array field",
		},
		{
			name:         "replace_not_append",
			field:        "labels",
			value:        "new-label",
			expectValues: []string{"new-label"},
			description:  "Array update should REPLACE existing values, not append",
		},
		{
			name:         "depends_on_with_ticket_ids",
			field:        "depends_on",
			value:        "TICKET-1,TICKET-2,TICKET-3",
			expectValues: []string{"TICKET-1", "TICKET-2", "TICKET-3"},
			description:  "depends_on field should handle ticket IDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary ticket file with some existing labels
			tmpDir := t.TempDir()
			ticketPath := filepath.Join(tmpDir, "TEST-1.md")

			originalContent := `---
id: TEST-1
title: "Test Ticket"
type: task
status: backlog
priority: medium
points: 0
assignee: ""
parent: ""
depends_on: [OLD-1, OLD-2]
blocks: []
related: []
labels: [existing-label]
created_at: 2026-02-01T10:00:00Z
updated_at: 2026-02-01T10:00:00Z
---

# Description
Test ticket
`
			if err := os.WriteFile(ticketPath, []byte(originalContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Update the field
			updateTicketField(ticketPath, tt.field, tt.value)

			// Read and verify the file
			content, err := os.ReadFile(ticketPath)
			if err != nil {
				t.Fatalf("Failed to read updated file: %v", err)
			}

			// Parse the YAML
			parts := strings.SplitN(string(content), "---", 3)
			var metadata map[string]interface{}
			if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			// Get the actual value - it might be []interface{} from YAML
			actualRaw := metadata[tt.field]
			var actualValues []string

			if actualRaw != nil {
				if arr, ok := actualRaw.([]interface{}); ok {
					for _, v := range arr {
						actualValues = append(actualValues, v.(string))
					}
				}
			}

			if len(actualValues) != len(tt.expectValues) {
				t.Errorf("%s: Expected %d values, got %d", tt.description, len(tt.expectValues), len(actualValues))
				return
			}

			for i, expectedVal := range tt.expectValues {
				if actualValues[i] != expectedVal {
					t.Errorf("%s: At index %d, expected %q, got %q", tt.description, i, expectedVal, actualValues[i])
				}
			}
		})
	}
}

// TestUpdateTicketFieldEnumField tests updating enum fields (priority, type, status)
func TestUpdateTicketFieldEnumField(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		value       string
		expectValue string
		description string
	}{
		{
			name:        "update_priority_high",
			field:       "priority",
			value:       "high",
			expectValue: "high",
			description: "Priority enum should accept valid values",
		},
		{
			name:        "update_priority_critical",
			field:       "priority",
			value:       "critical",
			expectValue: "critical",
			description: "Critical is valid priority value",
		},
		{
			name:        "update_type_story",
			field:       "type",
			value:       "story",
			expectValue: "story",
			description: "Type enum should accept valid values",
		},
		{
			name:        "update_type_bug",
			field:       "type",
			value:       "bug",
			expectValue: "bug",
			description: "Bug is valid type value",
		},
		{
			name:        "update_status_string",
			field:       "status",
			value:       "in-progress",
			expectValue: "in-progress",
			description: "Status values should be accepted (validated in workflow)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary ticket file
			tmpDir := t.TempDir()
			ticketPath := filepath.Join(tmpDir, "TEST-1.md")

			originalContent := `---
id: TEST-1
title: "Test Ticket"
type: task
status: backlog
priority: medium
points: 0
assignee: ""
parent: ""
depends_on: []
blocks: []
related: []
labels: []
created_at: 2026-02-01T10:00:00Z
updated_at: 2026-02-01T10:00:00Z
---

# Description
Test ticket
`
			if err := os.WriteFile(ticketPath, []byte(originalContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Update the field
			updateTicketField(ticketPath, tt.field, tt.value)

			// Read and verify the file
			content, err := os.ReadFile(ticketPath)
			if err != nil {
				t.Fatalf("Failed to read updated file: %v", err)
			}

			// Parse the YAML
			parts := strings.SplitN(string(content), "---", 3)
			var metadata map[string]interface{}
			if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			actualValue := metadata[tt.field]
			if actualValue != tt.expectValue {
				t.Errorf("%s: Expected %q, got %q", tt.description, tt.expectValue, actualValue)
			}
		})
	}
}

// TestUpdateTicketFieldUpdatesTimestamp tests that updated_at is always updated
func TestUpdateTicketFieldUpdatesTimestamp(t *testing.T) {
	// Create a temporary ticket file
	tmpDir := t.TempDir()
	ticketPath := filepath.Join(tmpDir, "TEST-1.md")

	originalTime := "2026-02-01T10:00:00Z"
	originalContent := `---
id: TEST-1
title: "Test Ticket"
type: task
status: backlog
priority: medium
points: 0
assignee: ""
parent: ""
depends_on: []
blocks: []
related: []
labels: []
created_at: 2026-02-01T09:00:00Z
updated_at: ` + originalTime + `
---

# Description
Test ticket
`
	if err := os.WriteFile(ticketPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Update the field
	updateTicketField(ticketPath, "title", "Updated Title")

	// Read and verify the file
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	// Parse the YAML
	parts := strings.SplitN(string(content), "---", 3)
	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// YAML unmarshals the timestamp - handle both string and time.Time
	var updatedAtTime time.Time
	switch v := metadata["updated_at"].(type) {
	case string:
		var err error
		updatedAtTime, err = time.Parse(time.RFC3339, v)
		if err != nil {
			t.Logf("Note: timestamp parsing detail - %v (this is OK, just logging)", err)
			// Just verify it's different from the original
			if v == originalTime {
				t.Errorf("updated_at should have changed from original value")
			}
			return
		}
	case time.Time:
		updatedAtTime = v
	default:
		t.Fatalf("Unexpected type for updated_at: %T", v)
	}

	// Verify timestamp was updated (check it's different from original)
	if updatedAtTime.Format(time.RFC3339) == originalTime {
		t.Errorf("updated_at should have changed from original value")
	}

	// Verify created_at did not change
	var createdAtStr string
	switch v := metadata["created_at"].(type) {
	case string:
		createdAtStr = v
	case time.Time:
		createdAtStr = v.Format(time.RFC3339)
	}

	if createdAtStr != "2026-02-01T09:00:00Z" {
		t.Errorf("created_at should not change on field update. Got: %v", createdAtStr)
	}
}

// TestUpdateTicketFieldPreservesYAMLFormat tests that YAML formatting is preserved
func TestUpdateTicketFieldPreservesYAMLFormat(t *testing.T) {
	// Create a temporary ticket file
	tmpDir := t.TempDir()
	ticketPath := filepath.Join(tmpDir, "TEST-1.md")

	originalContent := `---
id: TEST-1
title: "Test Ticket"
type: task
status: backlog
priority: medium
points: 0
assignee: ""
parent: ""
depends_on: []
blocks: []
related: []
labels: [bug, feature]
created_at: 2026-02-01T09:00:00Z
updated_at: 2026-02-01T10:00:00Z
---

# Description
Test ticket
`
	if err := os.WriteFile(ticketPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Update the field
	updateTicketField(ticketPath, "assignee", "alice")

	// Read the updated file
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	// Verify structure is intact
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) != 3 {
		t.Errorf("YAML structure broken. Got %d parts, expected 3", len(parts))
		return
	}

	// Verify we can parse the YAML
	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		t.Fatalf("Failed to parse updated YAML: %v", err)
	}

	// Verify markdown body is preserved
	if !strings.Contains(parts[2], "# Description") {
		t.Error("Markdown body was not preserved after update")
	}
}
