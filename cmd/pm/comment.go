package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment <ticket-id>",
	Short: "Add a comment to a ticket",
	Long: `Add a comment to a ticket without editing the main ticket file.

Comments are stored as separate files in .pm/tickets/{TICKET-ID}/ directory,
allowing multiple people to comment simultaneously without merge conflicts.

Interactive mode (default): Opens an editor for comment composition.
Direct mode: Use -m flag to provide comment text directly.

Examples:
  pm comment GPM-123                    # Open editor for comment
  pm comment GPM-123 -m "Looks good"    # Add comment directly
  pm comment GPM-123 -m "Fixed" --author bob  # Override author
`,
	Args: cobra.ExactArgs(1),
	RunE: runComment,
}

var commentFlags struct {
	message string
	author  string
}

func init() {
	commentCmd.Flags().StringVarP(&commentFlags.message, "message", "m", "", "Comment text (skips editor if provided)")
	commentCmd.Flags().StringVarP(&commentFlags.author, "author", "a", "", "Author name (defaults to git user.name)")
	rootCmd.AddCommand(commentCmd)
}

func runComment(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	pmDir := ".pm"

	// Verify ticket exists using case-insensitive lookup
	ticketPath := findTicketByID(ticketID)

	// Extract the actual ticket ID from the file path for use in comment directory
	actualTicketID := strings.TrimSuffix(filepath.Base(ticketPath), ".md")

	// Determine author
	author := commentFlags.author
	if author == "" {
		// Try to get author from git config
		gitAuthor, err := getGitAuthor()
		if err != nil {
			return fmt.Errorf("could not determine author: please set git user.name or use --author flag")
		}
		author = gitAuthor
	}

	var commentBody string

	if commentFlags.message != "" {
		// Direct mode: use provided message
		commentBody = commentFlags.message
	} else {
		// Interactive mode: open editor
		var err error
		commentBody, err = getCommentViaEditor(actualTicketID, author)
		if err != nil {
			return err
		}
	}

	// Validate comment is not empty
	if strings.TrimSpace(commentBody) == "" {
		return fmt.Errorf("empty comment not saved")
	}

	// Create comment file (use actual ticket ID from file, not user input)
	relPath, err := ticket.CreateCommentFile(actualTicketID, author, commentBody, pmDir)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	_ = relPath // Comment file created successfully

	fmt.Printf("✓ Comment added to %s\n", actualTicketID)
	return nil
}

// getGitAuthor retrieves the author name from git config user.name
func getGitAuthor() (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git config user.name failed: %w", err)
	}

	author := strings.TrimSpace(out.String())
	if author == "" {
		return "", fmt.Errorf("git user.name not configured")
	}

	return author, nil
}

// getCommentViaEditor opens the configured editor for comment composition
func getCommentViaEditor(ticketID string, author string) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "pm-comment-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write template with instructions
	template := fmt.Sprintf(`# Comment on %s by %s

# Please enter your comment above the line.
# Lines starting with # are ignored.
`, ticketID, author)

	if _, err := tmpFile.WriteString(template); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write template: %w", err)
	}
	tmpFile.Close()

	// Open editor
	editor := getEditor()
	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	// Read edited content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited comment: %w", err)
	}

	// Filter out comment lines and trim
	lines := strings.Split(string(content), "\n")
	var filteredLines []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			filteredLines = append(filteredLines, line)
		}
	}

	result := strings.TrimSpace(strings.Join(filteredLines, "\n"))
	return result, nil
}
