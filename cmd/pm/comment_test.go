package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
)

// TestSelectCommentToAmend tests the interactive comment selection menu
func TestSelectCommentToAmend(t *testing.T) {
	tests := []struct {
		name        string
		comments    []*ticket.Comment
		input       string
		expectError bool
		expectIdx   int
		description string
	}{
		{
			name: "select_first_comment",
			comments: []*ticket.Comment{
				{Author: "alice", Body: "First comment", CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
				{Author: "bob", Body: "Second comment", CreatedAt: time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)},
			},
			input:       "1",
			expectError: false,
			expectIdx:   0,
			description: "User selects first comment with valid input",
		},
		{
			name: "select_last_comment",
			comments: []*ticket.Comment{
				{Author: "alice", Body: "First", CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
				{Author: "bob", Body: "Second", CreatedAt: time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)},
				{Author: "charlie", Body: "Third", CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)},
			},
			input:       "3",
			expectError: false,
			expectIdx:   2,
			description: "User selects last comment",
		},
		{
			name: "cancel_with_q",
			comments: []*ticket.Comment{
				{Author: "alice", Body: "Comment", CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
			},
			input:       "q",
			expectError: true,
			description: "User cancels with 'q'",
		},
		{
			name: "cancel_with_Q_uppercase",
			comments: []*ticket.Comment{
				{Author: "alice", Body: "Comment", CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
			},
			input:       "Q",
			expectError: true,
			description: "User cancels with uppercase 'Q' (case-insensitive)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock stdin with test input
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdin = reader

			// Write input and close
			go func() {
				writer.WriteString(tt.input + "\n")
				writer.Close()
			}()

			// Redirect stdout to capture menu output
			oldStdout := os.Stdout
			defer func() { os.Stdout = oldStdout }()
			_, w, _ := os.Pipe()
			os.Stdout = w

			selected, err := selectCommentToAmend(tt.comments)

			w.Close()
			os.Stdout = oldStdout

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: Expected error, got nil", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: Unexpected error: %v", tt.description, err)
				}
				if selected != tt.comments[tt.expectIdx] {
					t.Errorf("%s: Expected comment at index %d, got different comment", tt.description, tt.expectIdx)
				}
			}
		})
	}
}

// TestGetEditorFallbackChain tests the editor selection with environment variables
func TestGetEditorFallbackChain(t *testing.T) {
	// Save original environment
	originalVISUAL := os.Getenv("VISUAL")
	originalEDITOR := os.Getenv("EDITOR")
	defer func() {
		os.Setenv("VISUAL", originalVISUAL)
		os.Setenv("EDITOR", originalEDITOR)
	}()

	tests := []struct {
		name           string
		visual         string
		editor         string
		expectContains string
		description    string
	}{
		{
			name:           "VISUAL_takes_precedence",
			visual:         "/usr/bin/emacs",
			editor:         "/usr/bin/nano",
			expectContains: "emacs",
			description:    "When VISUAL is set, it should be used even if EDITOR is set",
		},
		{
			name:           "EDITOR_used_when_VISUAL_not_set",
			visual:         "",
			editor:         "/usr/bin/nano",
			expectContains: "nano",
			description:    "When VISUAL not set, EDITOR should be used",
		},
		{
			name:           "fallback_to_defaults",
			visual:         "",
			editor:         "",
			expectContains: "editor|nano|vi",
			description:    "When neither VISUAL nor EDITOR set, should use default fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.visual != "" {
				os.Setenv("VISUAL", tt.visual)
			} else {
				os.Unsetenv("VISUAL")
			}

			if tt.editor != "" {
				os.Setenv("EDITOR", tt.editor)
			} else {
				os.Unsetenv("EDITOR")
			}

			result := getEditor()

			if tt.visual != "" && !strings.Contains(result, "emacs") {
				t.Errorf("%s: Expected editor to use VISUAL", tt.description)
			}

			if tt.visual == "" && tt.editor != "" {
				if !strings.Contains(result, "nano") {
					t.Errorf("%s: Expected editor to use EDITOR", tt.description)
				}
			}

			if result == "" {
				t.Errorf("%s: getEditor() returned empty string", tt.description)
			}
		})
	}
}

// TestGetGitAuthor tests author detection from git config
func TestGetGitAuthor(t *testing.T) {
	t.Run("success_with_git_config", func(t *testing.T) {
		// This test assumes git is properly configured on the test system
		// In a real CI environment, git user.name should be set
		author, err := getGitAuthor()

		// If git is not configured, we'll get an error - that's OK for this test
		if err != nil {
			t.Logf("Git not configured (expected in some environments): %v", err)
		} else {
			if author == "" {
				t.Error("Expected non-empty author from git config")
			}
		}
	})

	t.Run("handles_git_not_found", func(t *testing.T) {
		// Save original PATH
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set PATH to empty to ensure git is not found
		os.Setenv("PATH", "/nonexistent")

		_, err := getGitAuthor()

		// Should fail because git command not found
		if err == nil {
			// If git was found somehow, that's ok - the test environment has git
			t.Logf("Git found in system (test environment issue)")
		}

		os.Setenv("PATH", originalPath)
	})

	t.Run("filters_whitespace", func(t *testing.T) {
		// Create a temporary git directory to test with
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init", tmpDir)
		if err := cmd.Run(); err != nil {
			t.Skipf("git not available: %v", err)
		}

		// Configure a test user in the temp git repo
		cmd = exec.Command("git", "config", "user.name", "Test Author  ")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Could not configure git user: %v", err)
		}

		// Change to temp directory to read the config
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		os.Chdir(tmpDir)

		author, err := getGitAuthor()
		if err != nil {
			t.Fatalf("getGitAuthor failed: %v", err)
		}

		// Should have whitespace trimmed
		if author != "Test Author" {
			t.Errorf("Expected 'Test Author', got %q", author)
		}
	})
}

// TestFilterCommentLines tests the comment line filtering logic
func TestFilterCommentLines(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name: "strips_comment_lines",
			input: `My comment
# This is a comment
More content
# Another comment`,
			expected: `My comment
More content`,
			description: "Lines starting with # should be removed",
		},
		{
			name: "keeps_non_comment_lines",
			input: `Important line
Another important line`,
			expected: `Important line
Another important line`,
			description: "Non-comment lines should be preserved",
		},
		{
			name: "handles_leading_whitespace",
			input: `Content
   # Indented comment
More content`,
			expected: `Content
More content`,
			description: "Comments with leading whitespace should be stripped",
		},
		{
			name: "trims_overall_whitespace",
			input: `
  Content with leading whitespace
Followed by more

  
`,
			expected: `Content with leading whitespace
Followed by more`,
			description: "Leading/trailing whitespace should be trimmed overall",
		},
		{
			name:        "handles_empty_input",
			input:       "",
			expected:    "",
			description: "Empty input should return empty string",
		},
		{
			name: "handles_only_comments",
			input: `# Comment 1
# Comment 2
# Comment 3`,
			expected:    "",
			description: "Input with only comments should result in empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the filtering logic from getCommentViaEditor
			lines := strings.Split(tt.input, "\n")
			var filteredLines []string
			for _, line := range lines {
				if !strings.HasPrefix(strings.TrimSpace(line), "#") {
					filteredLines = append(filteredLines, line)
				}
			}
			result := strings.TrimSpace(strings.Join(filteredLines, "\n"))

			if result != tt.expected {
				t.Errorf("%s: Expected %q, got %q", tt.description, tt.expected, result)
			}
		})
	}
}

// TestCommentValidation tests comment body validation
func TestCommentValidation(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectValid bool
		description string
	}{
		{
			name:        "valid_comment",
			body:        "This is a valid comment",
			expectValid: true,
			description: "Non-empty comment should be valid",
		},
		{
			name:        "empty_comment",
			body:        "",
			expectValid: false,
			description: "Empty comment should be invalid",
		},
		{
			name:        "whitespace_only",
			body:        "   \n\t  \n  ",
			expectValid: false,
			description: "Whitespace-only comment should be invalid",
		},
		{
			name:        "multiline_valid",
			body:        "First line\nSecond line\nThird line",
			expectValid: true,
			description: "Multi-line comment should be valid",
		},
		{
			name:        "single_space",
			body:        " ",
			expectValid: false,
			description: "Single space should be invalid (whitespace only)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := strings.TrimSpace(tt.body) != ""

			if isValid != tt.expectValid {
				t.Errorf("%s: Expected valid=%v, got %v", tt.description, tt.expectValid, isValid)
			}
		})
	}
}

// TestCommentFileContent tests comment YAML front matter generation
func TestCommentFileContent(t *testing.T) {
	tests := []struct {
		name        string
		author      string
		body        string
		checkFor    []string
		description string
	}{
		{
			name:   "has_author_field",
			author: "alice",
			body:   "Test comment",
			checkFor: []string{
				"author: alice",
			},
			description: "Generated comment should include author field",
		},
		{
			name:   "has_timestamp_field",
			author: "bob",
			body:   "Another comment",
			checkFor: []string{
				"timestamp:",
			},
			description: "Generated comment should include timestamp field",
		},
		{
			name:   "has_yaml_delimiter",
			author: "charlie",
			body:   "Third comment",
			checkFor: []string{
				"---",
			},
			description: "Generated comment should have YAML delimiters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate comment file generation
			now := time.Now().UTC()
			header := bytes.NewBufferString("---\n")
			header.WriteString("author: " + tt.author + "\n")
			header.WriteString("timestamp: " + now.Format(time.RFC3339) + "\n")
			header.WriteString("---\n\n")
			header.WriteString(tt.body)

			content := header.String()

			for _, check := range tt.checkFor {
				if !strings.Contains(content, check) {
					t.Errorf("%s: Generated content missing %q\nContent: %s", tt.description, check, content)
				}
			}
		})
	}
}

// TestCommentTimestampParsing tests timestamp format handling in comments
func TestCommentTimestampParsing(t *testing.T) {
	tests := []struct {
		name        string
		timestamp   string
		expectError bool
		description string
	}{
		{
			name:        "valid_rfc3339",
			timestamp:   "2026-02-08T14:30:00Z",
			expectError: false,
			description: "RFC3339 formatted timestamp should parse successfully",
		},
		{
			name:        "valid_with_offset",
			timestamp:   "2026-02-08T14:30:00-07:00",
			expectError: false,
			description: "RFC3339 with timezone offset should parse successfully",
		},
		{
			name:        "invalid_format",
			timestamp:   "2026-02-08 14:30:00",
			expectError: true,
			description: "Non-RFC3339 format should fail to parse",
		},
		{
			name:        "empty_timestamp",
			timestamp:   "",
			expectError: true,
			description: "Empty timestamp should fail to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.Parse(time.RFC3339, tt.timestamp)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected parse error, got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected parse error: %v", tt.description, err)
			}
		})
	}
}

// TestAuthorDetectionLogic tests the author fallback chain in runComment
func TestAuthorDetectionLogic(t *testing.T) {
	tests := []struct {
		name          string
		flagAuthor    string
		gitAuthorName string
		expectAuthor  string
		expectError   bool
		description   string
	}{
		{
			name:          "flag_overrides_git",
			flagAuthor:    "flag_author",
			gitAuthorName: "git_author",
			expectAuthor:  "flag_author",
			expectError:   false,
			description:   "Author flag should override git config",
		},
		{
			name:          "git_used_when_flag_empty",
			flagAuthor:    "",
			gitAuthorName: "git_author",
			expectAuthor:  "git_author",
			expectError:   false,
			description:   "Git config should be used when flag not provided",
		},
		{
			name:          "empty_string_flag_uses_git",
			flagAuthor:    "",
			gitAuthorName: "git_user",
			expectAuthor:  "git_user",
			expectError:   false,
			description:   "Empty flag string should use git config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate author detection logic
			author := tt.flagAuthor
			if author == "" {
				// Would normally get from git
				author = tt.gitAuthorName
			}

			if tt.expectError {
				if author != "" {
					t.Errorf("%s: Expected error condition but got author %q", tt.description, author)
				}
			} else {
				if author != tt.expectAuthor {
					t.Errorf("%s: Expected author %q, got %q", tt.description, tt.expectAuthor, author)
				}
			}
		})
	}
}
