package milestone

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateID(t *testing.T) {
	valid := []string{"v1-0", "sprint-1", "mvp-launch", "v1-0-release"}
	for _, id := range valid {
		if err := ValidateID(id); err != nil {
			t.Errorf("ValidateID(%q) unexpected error: %v", id, err)
		}
	}

	invalid := []string{"v1.0", "Sprint 1", "_sprint", "V1", ""}
	for _, id := range invalid {
		if err := ValidateID(id); err == nil {
			t.Errorf("ValidateID(%q) expected error, got nil", id)
		}
	}
}

func TestSlugFromTitle(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"Version 1.0 Release", "version-1-0-release"},
		{"Sprint 3", "sprint-3"},
		{"My Feature!", "my-feature"},
		{"", "milestone"},
		{"---", "milestone"},
	}
	for _, c := range cases {
		got := SlugFromTitle(c.title)
		if got != c.want {
			t.Errorf("SlugFromTitle(%q) = %q, want %q", c.title, got, c.want)
		}
	}
}

func TestParseWriteRoundtrip(t *testing.T) {
	m := &Milestone{
		ID:        "v1-release",
		Title:     "Version 1 Release",
		State:     "active",
		CreatedAt: "2026-01-15T10:00:00Z",
		DueDate:   "2026-03-01",
		Body:      "# Notes\n\nSome body text.",
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "v1-release.md")

	if err := Write(m, path); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	content, _ := os.ReadFile(path)
	t.Logf("written file:\n%s", content)

	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	if got.ID != m.ID {
		t.Errorf("ID: got %q, want %q", got.ID, m.ID)
	}
	if got.Title != m.Title {
		t.Errorf("Title: got %q, want %q", got.Title, m.Title)
	}
	if got.State != m.State {
		t.Errorf("State: got %q, want %q", got.State, m.State)
	}
	if got.CreatedAt != m.CreatedAt {
		t.Errorf("CreatedAt: got %q, want %q", got.CreatedAt, m.CreatedAt)
	}
	if got.DueDate != m.DueDate {
		t.Errorf("DueDate: got %q, want %q", got.DueDate, m.DueDate)
	}
	if got.Body != m.Body {
		t.Errorf("Body: got %q, want %q", got.Body, m.Body)
	}
}

func TestValidate(t *testing.T) {
	valid := &Milestone{
		ID:        "v1-release",
		Title:     "Version 1",
		State:     "active",
		CreatedAt: "2026-01-15T10:00:00Z",
	}
	if err := Validate(valid); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Missing id
	m := *valid
	m.ID = ""
	if err := Validate(&m); err == nil {
		t.Error("expected error for missing id")
	}

	// Invalid state
	m = *valid
	m.State = "pending"
	if err := Validate(&m); err == nil {
		t.Error("expected error for invalid state")
	}

	// Closed state is valid
	m = *valid
	m.State = "closed"
	if err := Validate(&m); err != nil {
		t.Errorf("unexpected error for closed state: %v", err)
	}
}
