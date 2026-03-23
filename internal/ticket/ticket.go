package ticket

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Ticket struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Type      string   `yaml:"type"`
	Status    string   `yaml:"status"`
	Priority  string   `yaml:"priority,omitempty"`
	Points    int      `yaml:"points,omitempty"`
	Parent    string   `yaml:"parent,omitempty"`
	DependsOn []string `yaml:"depends_on,omitempty"`
	Blocks    []string `yaml:"blocks,omitempty"`
	Related   []string `yaml:"related,omitempty"`
	Labels     []string `yaml:"labels,omitempty"`
	Milestones []string `yaml:"milestones,omitempty"`
	Assignee   string   `yaml:"assignee,omitempty"`
	CreatedAt string   `yaml:"created_at"`
	UpdatedAt string   `yaml:"updated_at"`
	Body      string   `yaml:"-"`
}

var ticketIDPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*-\d+$`)

func (t *Ticket) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("id is required")
	}
	if !ticketIDPattern.MatchString(t.ID) {
		return fmt.Errorf("invalid id format: %s (expected PREFIX-123)", t.ID)
	}
	if t.Title == "" {
		return fmt.Errorf("title is required")
	}
	if t.Type == "" {
		return fmt.Errorf("type is required")
	}
	if t.Status == "" {
		return fmt.Errorf("status is required")
	}
	if _, err := time.Parse(time.RFC3339, t.CreatedAt); err != nil {
		return fmt.Errorf("invalid created_at format: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, t.UpdatedAt); err != nil {
		return fmt.Errorf("invalid updated_at format: %w", err)
	}
	return nil
}

// Normalize de-duplicates all array fields to maintain data integrity.
// Silent normalization - no errors, just cleans the data.
// Should be called before every save operation.
func (t *Ticket) Normalize() {
	t.DependsOn = dedup(t.DependsOn)
	t.Blocks = dedup(t.Blocks)
	t.Related = dedup(t.Related)
	t.Labels      = dedup(t.Labels)
	t.Milestones  = dedup(t.Milestones)
}

// dedup removes duplicate strings from a slice while preserving order.
func dedup(items []string) []string {
	if len(items) == 0 {
		return items
	}
	seen := make(map[string]bool)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func Parse(content []byte) (*Ticket, error) {
	// Split YAML front matter from Markdown body
	// Expected format for markdown files:
	// ---
	// id: value
	// ...
	// ---
	// # Markdown content
	//
	// Also support raw YAML (for testing)

	var yamlContent []byte

	// Check if content starts with --- (front-matter format)
	if bytes.HasPrefix(bytes.TrimSpace(content), []byte("---")) {
		parts := bytes.SplitN(content, []byte("---"), 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid ticket format: missing front matter delimiters")
		}

		// parts[0] is empty (before first ---)
		// parts[1] is the YAML front matter
		// parts[2] is the Markdown body
		yamlContent = parts[1]
	} else {
		// Raw YAML format (no front-matter delimiters)
		yamlContent = content
	}

	var t Ticket
	if err := yaml.Unmarshal(yamlContent, &t); err != nil {
		return nil, fmt.Errorf("failed to parse YAML front matter: %w", err)
	}
	return &t, nil
}
