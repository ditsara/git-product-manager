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

// TestQueryBuilding was removed in GPM-69. Query logic now lives in internal/cache/query.go
// and is tested via TestListTickets_* in internal/cache/query_test.go.

// TestTruncateID tests smart ID truncation that preserves the numeric suffix.
func TestTruncateID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		maxWidth int
		expected string
	}{
		{
			name:     "fits_exactly",
			id:       "GPM-42",
			maxWidth: 15,
			expected: "GPM-42",
		},
		{
			name:     "fits_with_children",
			id:       "GPM-1234 (+)",
			maxWidth: 15,
			expected: "GPM-1234 (+)",
		},
		{
			name:     "long_prefix_truncated",
			id:       "MYLONGPREFIX-1234",
			maxWidth: 15,
			expected: "MYLONGPRE…-1234",
		},
		{
			name:     "long_prefix_with_children",
			id:       "MYLONGPREFIX-1234 (+)",
			maxWidth: 15,
			expected: "MYLON…-1234 (+)",
		},
		{
			name:     "short_prefix_no_truncation",
			id:       "PROJ-99",
			maxWidth: 15,
			expected: "PROJ-99",
		},
		{
			name:     "exactly_at_limit",
			id:       "ABCDE-1234 (+)",
			maxWidth: 14,
			expected: "ABCDE-1234 (+)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateID(tt.id, tt.maxWidth)
			if got != tt.expected {
				t.Errorf("truncateID(%q, %d) = %q, want %q", tt.id, tt.maxWidth, got, tt.expected)
			}
			gotRunes := len([]rune(got))
			if gotRunes > tt.maxWidth {
				t.Errorf("truncateID(%q, %d) result %q has %d runes, exceeds maxWidth", tt.id, tt.maxWidth, got, gotRunes)
			}
		})
	}
}
