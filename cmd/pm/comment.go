package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/cache"
	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:               "comment <ticket-id>",
	Short:             "Add a comment to a ticket",
	ValidArgsFunction: completeTicketIDs,
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
	amend   bool
	timeStr string
}

func init() {
	commentCmd.Flags().StringVarP(&commentFlags.message, "message", "m", "", "Comment text (skips editor if provided)")
	commentCmd.Flags().StringVarP(&commentFlags.author, "author", "a", "", "Author name (defaults to git user.name)")
	commentCmd.Flags().BoolVar(&commentFlags.amend, "amend", false, "Edit an existing comment instead of creating a new one")
	commentCmd.Flags().StringVar(&commentFlags.timeStr, "timestamp", "", "Specific comment timestamp to edit (RFC3339 format)")
	rootCmd.AddCommand(commentCmd)
}

func runComment(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	pmDir := ".pm"

	// Verify ticket exists using case-insensitive lookup
	ticketPath := getTicketPath(ticketID)

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

	if commentFlags.amend {
		return amendComment(actualTicketID, author, pmDir)
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

func amendComment(ticketID string, author string, pmDir string) error {
	comments, err := ticket.ListCommentsForTicket(ticketID, pmDir)
	if err != nil {
		return err
	}

	if len(comments) == 0 {
		return fmt.Errorf("no comments found for %s", ticketID)
	}

	filtered := comments
	if author != "" {
		filtered = []*ticket.Comment{}
		for _, c := range comments {
			if c.Author == author {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("no comments by %s", author)
		}
	}

	var target *ticket.Comment
	if commentFlags.timeStr != "" {
		targetTime, err := time.Parse(time.RFC3339, commentFlags.timeStr)
		if err != nil {
			return fmt.Errorf("invalid timestamp %q: %w", commentFlags.timeStr, err)
		}

		for _, c := range filtered {
			if c.CreatedAt.Equal(targetTime) {
				target = c
				break
			}
		}
		if target == nil {
			return fmt.Errorf("comment with timestamp %s not found", commentFlags.timeStr)
		}
	} else if commentFlags.message == "" && len(filtered) > 1 {
		selected, err := selectCommentToAmend(filtered)
		if err != nil {
			return err
		}
		target = selected
	} else {
		// Default: most recent comment
		target = filtered[len(filtered)-1]
	}

	var newBody string
	if commentFlags.message != "" {
		newBody = commentFlags.message
	} else {
		newBody, err = getCommentEditViaEditor(ticketID, target.Author, target.Body)
		if err != nil {
			return err
		}
	}

	if strings.TrimSpace(newBody) == "" {
		return fmt.Errorf("empty comment not saved")
	}

	if err := ticket.UpdateCommentFile(target.Path, newBody, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	if err := cache.SyncCache(pmDir); err != nil {
		return fmt.Errorf("failed to update cache: %w", err)
	}

	fmt.Printf("✓ Comment updated on %s\n", ticketID)
	return nil
}

func selectCommentToAmend(comments []*ticket.Comment) (*ticket.Comment, error) {
	fmt.Println("Comments:")
	for i, c := range comments {
		preview := strings.SplitN(strings.TrimSpace(c.Body), "\n", 2)[0]
		createdAt := c.CreatedAt.UTC().Format("2006-01-02 15:04")
		fmt.Printf("[%d] @%s (%s): %s\n", i+1, c.Author, createdAt, preview)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Select comment to edit [1-%d] (or 'q' to cancel): ", len(comments))
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read selection: %w", err)
		}
		input = strings.TrimSpace(input)
		if strings.EqualFold(input, "q") {
			return nil, fmt.Errorf("comment edit cancelled")
		}
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(comments) {
			fmt.Println("Invalid selection.")
			continue
		}
		return comments[idx-1], nil
	}
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

// getCommentEditViaEditor opens the configured editor for editing an existing comment
func getCommentEditViaEditor(ticketID string, author string, currentBody string) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "pm-comment-edit-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write current body and instructions
	template := fmt.Sprintf("%s\n\n# Editing comment on %s by %s\n# Lines starting with # are ignored.\n", currentBody, ticketID, author)

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
