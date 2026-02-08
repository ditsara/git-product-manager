package ticket

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestNormalize tests the deduplication of array fields
func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		ticket   Ticket
		expected Ticket
	}{
		{
			name: "dedup depends_on",
			ticket: Ticket{
				ID:        "TEST-1",
				DependsOn: []string{"TEST-2", "TEST-3", "TEST-2"},
			},
			expected: Ticket{
				ID:        "TEST-1",
				DependsOn: []string{"TEST-2", "TEST-3"},
			},
		},
		{
			name: "dedup multiple arrays",
			ticket: Ticket{
				ID:        "TEST-1",
				DependsOn: []string{"A", "B", "A"},
				Blocks:    []string{"C", "C", "D"},
				Related:   []string{"E", "E"},
			},
			expected: Ticket{
				ID:        "TEST-1",
				DependsOn: []string{"A", "B"},
				Blocks:    []string{"C", "D"},
				Related:   []string{"E"},
			},
		},
		{
			name: "empty arrays unchanged",
			ticket: Ticket{
				ID:        "TEST-1",
				DependsOn: []string{},
				Blocks:    []string{},
			},
			expected: Ticket{
				ID:        "TEST-1",
				DependsOn: []string{},
				Blocks:    []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ticket.Normalize()

			// Compare DependsOn
			if !sliceEqual(tt.ticket.DependsOn, tt.expected.DependsOn) {
				t.Errorf("DependsOn: got %v, want %v", tt.ticket.DependsOn, tt.expected.DependsOn)
			}
			if !sliceEqual(tt.ticket.Blocks, tt.expected.Blocks) {
				t.Errorf("Blocks: got %v, want %v", tt.ticket.Blocks, tt.expected.Blocks)
			}
			if !sliceEqual(tt.ticket.Related, tt.expected.Related) {
				t.Errorf("Related: got %v, want %v", tt.ticket.Related, tt.expected.Related)
			}
		})
	}
}

// TestUpdateRelationshipWithSymmetry tests linking tickets with symmetry
func TestUpdateRelationshipWithSymmetry(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	ticketsPath := filepath.Join(pmPath, "tickets")

	if err := os.MkdirAll(ticketsPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test tickets
	createTestTicket(t, ticketsPath, "TEST-1", "Ticket 1")
	createTestTicket(t, ticketsPath, "TEST-2", "Ticket 2")
	createTestTicket(t, ticketsPath, "TEST-3", "Ticket 3")

	tests := []struct {
		name         string
		sourceID     string
		targetID     string
		relType      string
		add          bool
		wantErr      bool
		expectExists bool
		verifyFunc   func(t *testing.T, source, target *Ticket)
	}{
		{
			name:         "add depends-on relationship",
			sourceID:     "TEST-1",
			targetID:     "TEST-2",
			relType:      "depends-on",
			add:          true,
			wantErr:      false,
			expectExists: false,
			verifyFunc: func(t *testing.T, source, target *Ticket) {
				if !contains(source.DependsOn, "TEST-2") {
					t.Errorf("TEST-1 should depend on TEST-2")
				}
				if !contains(target.Blocks, "TEST-1") {
					t.Errorf("TEST-2 should block TEST-1")
				}
			},
		},
		{
			name:         "add blocks relationship",
			sourceID:     "TEST-2",
			targetID:     "TEST-3",
			relType:      "blocks",
			add:          true,
			wantErr:      false,
			expectExists: false,
			verifyFunc: func(t *testing.T, source, target *Ticket) {
				if !contains(source.Blocks, "TEST-3") {
					t.Errorf("TEST-2 should block TEST-3")
				}
				if !contains(target.DependsOn, "TEST-2") {
					t.Errorf("TEST-3 should depend on TEST-2")
				}
			},
		},
		{
			name:         "add related link",
			sourceID:     "TEST-1",
			targetID:     "TEST-3",
			relType:      "related",
			add:          true,
			wantErr:      false,
			expectExists: false,
			verifyFunc: func(t *testing.T, source, target *Ticket) {
				if !contains(source.Related, "TEST-3") {
					t.Errorf("TEST-1 should be related to TEST-3")
				}
				if contains(target.Related, "TEST-1") {
					t.Errorf("TEST-3 should NOT be related to TEST-1 (unidirectional)")
				}
			},
		},
		{
			name:         "add duplicate (already exists)",
			sourceID:     "TEST-1",
			targetID:     "TEST-2",
			relType:      "depends-on",
			add:          true,
			wantErr:      false,
			expectExists: true,
			verifyFunc: func(t *testing.T, source, target *Ticket) {
				// Should not have duplicates
				count := 0
				for _, dep := range source.DependsOn {
					if dep == "TEST-2" {
						count++
					}
				}
				if count != 1 {
					t.Errorf("TEST-1 should depend on TEST-2 exactly once, got %d", count)
				}
			},
		},
		{
			name:         "self-reference error",
			sourceID:     "TEST-1",
			targetID:     "TEST-1",
			relType:      "depends-on",
			add:          true,
			wantErr:      true,
			expectExists: false,
		},
		{
			name:         "missing target error",
			sourceID:     "TEST-1",
			targetID:     "FAKE-999",
			relType:      "depends-on",
			add:          true,
			wantErr:      true,
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := UpdateRelationshipWithSymmetry(pmPath, tt.sourceID, tt.targetID, tt.relType, tt.add)

			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, want error %v", err, tt.wantErr)
			}

			if err == nil {
				if exists != tt.expectExists {
					t.Errorf("got exists=%v, want %v", exists, tt.expectExists)
				}

				if tt.verifyFunc != nil {
					// Load and verify tickets
					sourcePath := filepath.Join(ticketsPath, tt.sourceID+".md")
					targetPath := filepath.Join(ticketsPath, tt.targetID+".md")

					sourceContent, _ := os.ReadFile(sourcePath)
					targetContent, _ := os.ReadFile(targetPath)

					source, _ := Parse(sourceContent)
					target, _ := Parse(targetContent)

					tt.verifyFunc(t, source, target)
				}
			}
		})
	}
}

// TestRemoveRelationshipFromAllFields tests removing relationships
func TestRemoveRelationshipFromAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	ticketsPath := filepath.Join(pmPath, "tickets")

	if err := os.MkdirAll(ticketsPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test ticket with multiple relationships
	ticket := &Ticket{
		ID:        "TEST-1",
		Title:     "Test Ticket",
		Type:      "task",
		Status:    "backlog",
		DependsOn: []string{"TEST-2", "TEST-3"},
		Blocks:    []string{"TEST-4"},
		Related:   []string{"TEST-5"},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Create other tickets
	for id := 2; id <= 5; id++ {
		createTestTicket(t, ticketsPath, "TEST-"+string(rune(48+id)), "Ticket "+string(rune(48+id)))
	}

	// Save the main ticket
	content := serializeTicket(ticket, "")
	if err := os.WriteFile(filepath.Join(ticketsPath, "TEST-1.md"), content, 0644); err != nil {
		t.Fatalf("Failed to write test ticket: %v", err)
	}

	// Test removing TEST-2 from all fields
	wasNotLinked, err := RemoveRelationshipFromAllFields(pmPath, "TEST-1", "TEST-2")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if wasNotLinked {
		t.Errorf("Expected wasNotLinked=false, got true")
	}

	// Verify TEST-2 was removed from depends_on
	updatedContent, _ := os.ReadFile(filepath.Join(ticketsPath, "TEST-1.md"))
	updated, _ := Parse(updatedContent)

	if contains(updated.DependsOn, "TEST-2") {
		t.Errorf("TEST-2 should be removed from depends_on")
	}
	if !contains(updated.DependsOn, "TEST-3") {
		t.Errorf("TEST-3 should still be in depends_on")
	}
}

// Helper functions

func createTestTicket(t *testing.T, path, id, title string) {
	ticket := &Ticket{
		ID:        id,
		Title:     title,
		Type:      "task",
		Status:    "backlog",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	content := serializeTicket(ticket, "")
	if err := os.WriteFile(filepath.Join(path, id+".md"), content, 0644); err != nil {
		t.Fatalf("Failed to create test ticket: %v", err)
	}
}

func serializeTicket(ticket *Ticket, body string) []byte {
	yamlData, _ := yaml.Marshal(ticket)
	return []byte("---\n" + string(yamlData) + "---\n" + body)
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
