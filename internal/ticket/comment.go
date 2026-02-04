package ticket

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Comment represents a single comment on a ticket.
type Comment struct {
	Author    string    `yaml:"author"`
	Timestamp time.Time `yaml:"-"` // Stored as ISO8601 string in YAML
	Body      string    `yaml:"-"`
}

// CommentMetadata is the YAML front-matter of a comment file
type CommentMetadata struct {
	Author    string `yaml:"author"`
	Timestamp string `yaml:"timestamp"` // ISO8601 format
}

// ParseCommentFile reads and parses a comment file
// Returns the metadata and body content separately
func ParseCommentFile(filepath string) (*Comment, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read comment file: %w", err)
	}

	contentStr := string(content)

	// Split front-matter and body
	parts := strings.SplitN(contentStr, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid comment file format: missing front-matter delimiters")
	}

	// Parse YAML front-matter
	var metadata CommentMetadata
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse comment metadata: %w", err)
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, metadata.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format %q: %w", metadata.Timestamp, err)
	}

	// Extract body (trim leading/trailing whitespace)
	body := strings.TrimSpace(parts[2])

	return &Comment{
		Author:    metadata.Author,
		Timestamp: timestamp,
		Body:      body,
	}, nil
}

// GenerateCommentFilename creates a filesystem-safe filename for a comment
// Format: {ISO8601-timestamp}-{author}.md
// Example: 2026-02-03T14-30-00Z-alice.md
func GenerateCommentFilename(author string, timestamp time.Time) string {
	// Format timestamp with hyphens instead of colons
	isoTime := timestamp.UTC().Format("2006-01-02T15-04-05Z")
	// Sanitize author name (remove spaces, special chars)
	sanitizedAuthor := strings.ToLower(author)
	sanitizedAuthor = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, sanitizedAuthor)

	return fmt.Sprintf("%s-%s.md", isoTime, sanitizedAuthor)
}

// CreateCommentFile writes a new comment file
// Creates the ticket comment directory if it doesn't exist
// Returns the relative path to the created comment file
func CreateCommentFile(ticketID string, author string, body string, baseDir string) (string, error) {
	// Create comments directory if it doesn't exist
	commentDir := filepath.Join(baseDir, "tickets", ticketID)
	if err := os.MkdirAll(commentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create comment directory: %w", err)
	}

	// Generate filename and path
	now := time.Now().UTC()
	filename := GenerateCommentFilename(author, now)
	commentPath := filepath.Join(commentDir, filename)

	// Check if file already exists (same-second collision)
	// If so, add a counter
	counter := 1
	for {
		if _, err := os.Stat(commentPath); err != nil {
			if os.IsNotExist(err) {
				break // File doesn't exist, we can use this name
			}
			return "", fmt.Errorf("failed to check comment file existence: %w", err)
		}
		// File exists, try with counter
		parts := strings.SplitN(filename, ".md", 2)
		commentPath = filepath.Join(commentDir, fmt.Sprintf("%s-%d.md", parts[0], counter))
		counter++
	}

	// Build comment content
	metadata := CommentMetadata{
		Author:    author,
		Timestamp: now.Format(time.RFC3339),
	}

	metadataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal comment metadata: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n\n%s", string(metadataBytes), body)

	// Write comment file
	if err := os.WriteFile(commentPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write comment file: %w", err)
	}

	// Return relative path from .pm/tickets
	relPath := filepath.Join(ticketID, filepath.Base(commentPath))
	return relPath, nil
}

// ListCommentsForTicket returns all comments for a ticket, sorted by timestamp ascending
func ListCommentsForTicket(ticketID string, baseDir string) ([]*Comment, error) {
	commentDir := filepath.Join(baseDir, "tickets", ticketID)

	// Check if comment directory exists
	info, err := os.Stat(commentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Comment{}, nil // No comments yet
		}
		return nil, fmt.Errorf("failed to access comment directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("comment path is not a directory")
	}

	// Read all files in comment directory
	entries, err := os.ReadDir(commentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read comment directory: %w", err)
	}

	var comments []*Comment
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		commentPath := filepath.Join(commentDir, entry.Name())
		comment, err := ParseCommentFile(commentPath)
		if err != nil {
			// Log warning but continue - don't fail entire operation for one bad comment
			continue
		}

		comments = append(comments, comment)
	}

	// Sort by timestamp ascending
	for i := 0; i < len(comments)-1; i++ {
		for j := i + 1; j < len(comments); j++ {
			if comments[j].Timestamp.Before(comments[i].Timestamp) {
				comments[i], comments[j] = comments[j], comments[i]
			}
		}
	}

	return comments, nil
}
