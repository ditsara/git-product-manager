package ticket

import (
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
	Labels    []string `yaml:"labels,omitempty"`
	Assignee  string   `yaml:"assignee,omitempty"`
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

func Parse(content []byte) (*Ticket, error) {
	var t Ticket
	if err := yaml.Unmarshal(content, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
