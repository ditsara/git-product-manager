package ticket

import (
	"reflect"
	"testing"
)

func TestParseFieldValue_Arrays(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		input    string
		expected []string
	}{
		{
			name:     "multiple elements",
			field:    "labels",
			input:    "bug,critical,p1",
			expected: []string{"bug", "critical", "p1"},
		},
		{
			name:     "single element",
			field:    "labels",
			input:    "bug",
			expected: []string{"bug"},
		},
		{
			name:     "empty value",
			field:    "labels",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace trimming",
			field:    "labels",
			input:    " bug , critical , p1 ",
			expected: []string{"bug", "critical", "p1"},
		},
		{
			name:     "empty elements filtered",
			field:    "labels",
			input:    "bug,,critical",
			expected: []string{"bug", "critical"},
		},
		{
			name:     "semicolon treated as part of value",
			field:    "labels",
			input:    "bug;critical",
			expected: []string{"bug;critical"},
		},
		{
			name:     "pipe treated as part of value",
			field:    "labels",
			input:    "bug|critical",
			expected: []string{"bug|critical"},
		},
		{
			name:     "depends_on with ticket IDs",
			field:    "depends_on",
			input:    "GPM-1,GPM-2,GPM-3",
			expected: []string{"GPM-1", "GPM-2", "GPM-3"},
		},
		{
			name:     "blocks field",
			field:    "blocks",
			input:    "GPM-10,GPM-11",
			expected: []string{"GPM-10", "GPM-11"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFieldValue(tt.field, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultSlice, ok := result.([]string)
			if !ok {
				t.Fatalf("expected []string, got %T", result)
			}

			if !reflect.DeepEqual(resultSlice, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, resultSlice)
			}
		})
	}
}

func TestParseFieldValue_Integers(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		input       string
		expected    int
		expectError bool
	}{
		{
			name:     "valid integer",
			field:    "points",
			input:    "5",
			expected: 5,
		},
		{
			name:     "zero",
			field:    "points",
			input:    "0",
			expected: 0,
		},
		{
			name:     "empty string",
			field:    "points",
			input:    "",
			expected: 0,
		},
		{
			name:        "invalid integer",
			field:       "points",
			input:       "abc",
			expectError: true,
		},
		{
			name:        "float not allowed",
			field:       "points",
			input:       "5.5",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFieldValue(tt.field, tt.input)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultInt, ok := result.(int)
			if !ok {
				t.Fatalf("expected int, got %T", result)
			}

			if resultInt != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, resultInt)
			}
		})
	}
}

func TestParseFieldValue_Enums(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "valid type: epic",
			field:    "type",
			input:    "epic",
			expected: "epic",
		},
		{
			name:     "valid type: story",
			field:    "type",
			input:    "story",
			expected: "story",
		},
		{
			name:     "valid type: task",
			field:    "type",
			input:    "task",
			expected: "task",
		},
		{
			name:     "valid type: bug",
			field:    "type",
			input:    "bug",
			expected: "bug",
		},
		{
			name:        "invalid type",
			field:       "type",
			input:       "feature",
			expectError: true,
		},
		{
			name:     "valid priority: low",
			field:    "priority",
			input:    "low",
			expected: "low",
		},
		{
			name:     "valid priority: medium",
			field:    "priority",
			input:    "medium",
			expected: "medium",
		},
		{
			name:     "valid priority: high",
			field:    "priority",
			input:    "high",
			expected: "high",
		},
		{
			name:     "valid priority: critical",
			field:    "priority",
			input:    "critical",
			expected: "critical",
		},
		{
			name:        "invalid priority",
			field:       "priority",
			input:       "urgent",
			expectError: true,
		},
		{
			name:     "status (validated elsewhere)",
			field:    "status",
			input:    "backlog",
			expected: "backlog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFieldValue(tt.field, tt.input)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("expected string, got %T", result)
			}

			if resultStr != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resultStr)
			}
		})
	}
}

func TestParseFieldValue_Strings(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		input    string
		expected string
	}{
		{
			name:     "title",
			field:    "title",
			input:    "My Ticket Title",
			expected: "My Ticket Title",
		},
		{
			name:     "assignee",
			field:    "assignee",
			input:    "alice",
			expected: "alice",
		},
		{
			name:     "parent ticket ID",
			field:    "parent",
			input:    "GPM-1",
			expected: "GPM-1",
		},
		{
			name:     "empty assignee",
			field:    "assignee",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFieldValue(tt.field, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("expected string, got %T", result)
			}

			if resultStr != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resultStr)
			}
		})
	}
}

func TestParseFieldValue_UnknownField(t *testing.T) {
	// Unknown fields should be treated as strings
	result, err := ParseFieldValue("unknown_field", "some value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("expected string for unknown field, got %T", result)
	}

	if resultStr != "some value" {
		t.Errorf("expected 'some value', got '%s'", resultStr)
	}
}
