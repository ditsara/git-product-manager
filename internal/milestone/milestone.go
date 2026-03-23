package milestone

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Milestone struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	DueDate     string `yaml:"due_date,omitempty"`
	State       string `yaml:"state"`
	CreatedAt   string `yaml:"created_at"`
	ClosedAt    string `yaml:"closed_at,omitempty"`
	Body        string `yaml:"-"`
}

func Parse(content []byte) (*Milestone, error) {
	var yamlContent []byte
	var body string

	if bytes.HasPrefix(bytes.TrimSpace(content), []byte("---")) {
		parts := bytes.SplitN(content, []byte("---"), 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid milestone format: missing front matter delimiters")
		}
		yamlContent = parts[1]
		body = strings.TrimSpace(string(parts[2]))
	} else {
		yamlContent = content
	}

	var m Milestone
	if err := yaml.Unmarshal(yamlContent, &m); err != nil {
		return nil, fmt.Errorf("failed to parse YAML front matter: %w", err)
	}
	m.Body = body
	return &m, nil
}

func ParseFile(path string) (*Milestone, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read milestone file: %w", err)
	}
	return Parse(content)
}

func Write(m *Milestone, path string) error {
	yamlBytes, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal milestone: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n")
	if m.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(m.Body)
		buf.WriteString("\n")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write milestone file: %w", err)
	}
	return nil
}

func Validate(m *Milestone) error {
	if m.ID == "" {
		return fmt.Errorf("id is required")
	}
	if err := ValidateID(m.ID); err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	if m.Title == "" {
		return fmt.Errorf("title is required")
	}
	if m.State == "" {
		return fmt.Errorf("state is required")
	}
	if m.State != "active" && m.State != "closed" {
		return fmt.Errorf("invalid state %q: must be \"active\" or \"closed\"", m.State)
	}
	if m.CreatedAt == "" {
		return fmt.Errorf("created_at is required")
	}
	return nil
}

func ListMilestones(milestonesPath string) ([]*Milestone, error) {
	entries, err := os.ReadDir(milestonesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read milestones directory: %w", err)
	}

	var milestones []*Milestone
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(milestonesPath, entry.Name())
		m, err := ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}
		milestones = append(milestones, m)
	}
	return milestones, nil
}
