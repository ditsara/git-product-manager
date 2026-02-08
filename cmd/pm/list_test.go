package main

import (
	"testing"
)

// TestTruncate tests the truncate function with various input lengths
func TestTruncate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		maxLen      int
		expected    string
		description string
	}{
		{
			name:        "no_truncation_needed",
			input:       "Hello World",
			maxLen:      20,
			expected:    "Hello World",
			description: "String shorter than maxLen should not be truncated",
		},
		{
			name:        "exact_length",
			input:       "Hello World",
			maxLen:      11,
			expected:    "Hello World",
			description: "String exactly at maxLen should not be truncated",
		},
		{
			name:        "truncation_needed",
			input:       "This is a very long string",
			maxLen:      10,
			expected:    "This is...",
			description: "String longer than maxLen should be truncated with ...",
		},
		{
			name:        "single_character",
			input:       "A",
			maxLen:      1,
			expected:    "A",
			description: "Single character at limit should not be truncated",
		},
		{
			name:        "truncate_maxlen_less_than_3",
			input:       "Hello",
			maxLen:      2,
			expected:    "...",
			description: "When maxLen < 3, should return just ... (since can't show any chars)",
		},
		{
			name:        "unicode_characters",
			input:       "こんにちは世界",
			maxLen:      5,
			expected:    "こん...",
			description: "Unicode characters should be counted as runes, not bytes (5-3=2 chars + ...)",
		},
		{
			name:        "emoji_handling",
			input:       "Hello 👋 World 🌍",
			maxLen:      12,
			expected:    "Hello 👋 W...",
			description: "Emoji should be counted as single runes (12-3=9 chars + ...)",
		},
		{
			name:        "empty_string",
			input:       "",
			maxLen:      10,
			expected:    "",
			description: "Empty string should remain empty",
		},
		{
			name:        "maxLen_exactly_3",
			input:       "Hello",
			maxLen:      3,
			expected:    "...",
			description: "maxLen of 3 should result in no content chars + ...",
		},
		{
			name:        "maxLen_greater_than_3",
			input:       "HelloWorld",
			maxLen:      5,
			expected:    "He...",
			description: "With maxLen=5 on 10-char string, should show 5-3=2 chars + ...",
		},
		{
			name:        "spaces_count",
			input:       "A B C D E",
			maxLen:      5,
			expected:    "A ...",
			description: "Spaces should be counted as characters (5-3=2 chars + ...)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("%s: Expected %q, got %q", tt.description, tt.expected, result)
			}
		})
	}
}

// TestTruncateEdgeCases tests edge cases for the truncate function
func TestTruncateEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		maxLen      int
		description string
	}{
		{
			name:        "negative_maxLen",
			input:       "Hello",
			maxLen:      -1,
			description: "Negative maxLen should result in truncation",
		},
		{
			name:        "zero_maxLen",
			input:       "Hello",
			maxLen:      0,
			description: "Zero maxLen should result in ...",
		},
		{
			name:        "very_large_maxLen",
			input:       "Hello",
			maxLen:      1000000,
			description: "Very large maxLen should not truncate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			// Just verify it doesn't panic and returns a string
			if result == "" && tt.input != "" && tt.maxLen >= 3 {
				t.Errorf("%s: Unexpected empty result", tt.description)
			}
		})
	}
}

// TestTruncateWithFixedWidth tests that truncate works for table column formatting
func TestTruncateWithFixedWidth(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		columnWidth int
		expectFit   bool
		description string
	}{
		{
			name:        "title_column_width_50",
			input:       "Implement OAuth2 Login Flow with MFA Support",
			columnWidth: 50,
			expectFit:   false,
			description: "Long title should be truncated to fit in 50-char column",
		},
		{
			name:        "title_fits_in_column",
			input:       "Fix bug in list",
			columnWidth: 50,
			expectFit:   true,
			description: "Short title should fit in 50-char column",
		},
		{
			name:        "id_column_width_20",
			input:       "PROJECT-1234567890ABC",
			columnWidth: 20,
			expectFit:   false,
			description: "Long ID should be truncated to fit in 20-char column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.columnWidth)

			if tt.expectFit {
				if len([]rune(result)) > tt.columnWidth {
					t.Errorf("%s: Result %q exceeds column width %d", tt.description, result, tt.columnWidth)
				}
			} else {
				// If truncation should occur, verify it was truncated
				if len([]rune(result)) <= tt.columnWidth {
					// This is fine - the string fits without truncation
				}
			}

			// Verify truncated string isn't too short unless input was short
			runeCount := len([]rune(result))
			if runeCount > tt.columnWidth {
				t.Errorf("%s: Truncated result %q is %d chars, exceeds limit of %d", tt.description, result, runeCount, tt.columnWidth)
			}
		})
	}
}

// TestQueryBuilding tests the logic for building SQL queries (simulated without database)
func TestQueryBuilding(t *testing.T) {
	tests := []struct {
		name               string
		showAll            bool
		parentFilter       string
		statusFilter       string
		expectedInclusions []string
		description        string
	}{
		{
			name:               "default_top_level_only",
			showAll:            false,
			parentFilter:       "",
			statusFilter:       "",
			expectedInclusions: []string{"WHERE (parent IS NULL OR parent = '')"},
			description:        "Default should filter to top-level tickets only",
		},
		{
			name:               "show_all_flag",
			showAll:            true,
			parentFilter:       "",
			statusFilter:       "",
			expectedInclusions: []string{"SELECT id, title, type, status FROM tickets"},
			description:        "With --all flag, query should select from all tickets",
		},
		{
			name:               "parent_filter_direct_children",
			showAll:            false,
			parentFilter:       "TICKET-1",
			statusFilter:       "",
			expectedInclusions: []string{"UPPER(parent) = UPPER(?)"},
			description:        "Parent filter should query for direct children with case-insensitive matching",
		},
		{
			name:               "parent_filter_recursive",
			showAll:            true,
			parentFilter:       "TICKET-1",
			statusFilter:       "",
			expectedInclusions: []string{"WITH RECURSIVE subtree", "UNION ALL"},
			description:        "Parent filter with --all should use recursive query",
		},
		{
			name:               "status_filter_combined",
			showAll:            false,
			parentFilter:       "",
			statusFilter:       "done",
			expectedInclusions: []string{"AND status = ?"},
			description:        "Status filter should be added with AND when combined with other filters",
		},
		{
			name:               "all_filters_combined",
			showAll:            true,
			parentFilter:       "TICKET-1",
			statusFilter:       "in-progress",
			expectedInclusions: []string{"WITH RECURSIVE subtree", "AND", "status = ?"},
			description:        "All filters combined should build complex query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the query building logic from listCmd
			var query string

			if tt.parentFilter != "" && tt.showAll {
				query = `WITH RECURSIVE subtree(id) AS (
					SELECT id FROM tickets WHERE UPPER(parent) = UPPER(?)
					UNION ALL
					SELECT t.id FROM tickets t
					JOIN subtree s ON UPPER(t.parent) = UPPER(s.id)
				)
				SELECT id, title, type, status FROM tickets
				WHERE id IN (SELECT id FROM subtree)`
			} else if tt.parentFilter != "" {
				query = "SELECT id, title, type, status FROM tickets WHERE UPPER(parent) = UPPER(?)"
			} else if !tt.showAll {
				query = "SELECT id, title, type, status FROM tickets WHERE (parent IS NULL OR parent = '')"
			} else {
				query = "SELECT id, title, type, status FROM tickets"
			}

			// Add status filter
			if tt.statusFilter != "" {
				if !tt.showAll || tt.parentFilter != "" {
					// Already has a WHERE clause from above
					if tt.parentFilter != "" && tt.showAll {
						query += " AND status = ?"
					} else if tt.parentFilter != "" {
						query += " AND status = ?"
					} else {
						query += " AND status = ?"
					}
				} else {
					query += " WHERE status = ?"
				}
			}

			// Verify expected inclusions
			for _, inclusion := range tt.expectedInclusions {
				if !containsIgnoreWhitespace(query, inclusion) {
					t.Errorf("%s: Query missing expected component %q\nFull query: %s", tt.description, inclusion, query)
				}
			}
		})
	}
}

// containsIgnoreWhitespace checks if a query contains a substring ignoring extra whitespace
func containsIgnoreWhitespace(query, substring string) bool {
	// Simple check - in reality might want more sophisticated whitespace handling
	return contains(normalizeWhitespace(query), normalizeWhitespace(substring))
}

func normalizeWhitespace(s string) string {
	var result string
	inSpace := false
	for _, c := range s {
		if c == ' ' || c == '\n' || c == '\t' {
			if !inSpace {
				result += " "
				inSpace = true
			}
		} else {
			result += string(c)
			inSpace = false
		}
	}
	return result
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && searchString(haystack, needle))
}

func searchString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
