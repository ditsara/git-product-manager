package guide

import (
	"fmt"
	"strings"
)

// sectionOrder controls the sequence of sections in full output.
var sectionOrder = []string{"workflow", "schema", "commands", "principles"}

// sectionFiles maps section names to their embedded .md filenames.
var sectionFiles = map[string]string{
	"workflow":   "workflow.md",
	"schema":     "schema.md",
	"commands":   "commands.md",
	"principles": "principles.md",
}

// SectionNames returns the ordered list of available section names.
func SectionNames() []string {
	names := make([]string, len(sectionOrder))
	copy(names, sectionOrder)
	return names
}

// Section returns the Markdown content for a single named section.
func Section(name string) (string, error) {
	filename, ok := sectionFiles[name]
	if !ok {
		return "", fmt.Errorf("unknown section %q — valid sections: %s",
			name, strings.Join(sectionOrder, ", "))
	}
	data, err := FS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read guide section %q: %w", name, err)
	}
	return string(data), nil
}

// Full returns all sections concatenated in order, separated by a horizontal rule.
func Full() (string, error) {
	var parts []string
	for _, name := range sectionOrder {
		content, err := Section(name)
		if err != nil {
			return "", err
		}
		parts = append(parts, strings.TrimSpace(content))
	}
	return strings.Join(parts, "\n\n---\n\n") + "\n", nil
}
