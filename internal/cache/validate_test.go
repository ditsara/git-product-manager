package cache

import (
	"strings"
	"testing"
)

func TestValidateRelationshipSymmetry(t *testing.T) {
	t.Run("clean_data_no_warnings", func(t *testing.T) {
		rels := []relationshipData{
			{fromTicket: "A-1", toTicket: "A-2", relType: "depends-on"},
			{fromTicket: "A-2", toTicket: "A-1", relType: "blocks"},
			{fromTicket: "A-3", toTicket: "A-4", relType: "related"},
			{fromTicket: "A-4", toTicket: "A-3", relType: "related"},
		}
		warnings := validateRelationshipSymmetry(rels)
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %d: %v", len(warnings), warnings)
		}
	})

	t.Run("missing_blocks_for_depends_on", func(t *testing.T) {
		rels := []relationshipData{
			{fromTicket: "A-5", toTicket: "A-10", relType: "depends-on"},
			// A-10 does NOT have blocks: [A-5]
		}
		warnings := validateRelationshipSymmetry(rels)
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		w := warnings[0]
		if !strings.Contains(w, "A-5 depends_on A-10") {
			t.Errorf("warning missing problem description: %s", w)
		}
		if !strings.Contains(w, "pm link A-10 A-5 --type blocks") {
			t.Errorf("warning missing fix command: %s", w)
		}
	})

	t.Run("missing_depends_on_for_blocks", func(t *testing.T) {
		rels := []relationshipData{
			{fromTicket: "A-999", toTicket: "A-1000", relType: "blocks"},
			// A-1000 does NOT have depends_on: [A-999]
		}
		warnings := validateRelationshipSymmetry(rels)
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		w := warnings[0]
		if !strings.Contains(w, "A-999 blocks A-1000") {
			t.Errorf("warning missing problem description: %s", w)
		}
		if !strings.Contains(w, "pm link A-1000 A-999 --type depends-on") {
			t.Errorf("warning missing fix command: %s", w)
		}
	})

	t.Run("missing_reverse_related", func(t *testing.T) {
		rels := []relationshipData{
			{fromTicket: "A-3", toTicket: "A-7", relType: "related"},
			// A-7 does NOT have related: [A-3]
		}
		warnings := validateRelationshipSymmetry(rels)
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		w := warnings[0]
		if !strings.Contains(w, "A-3 related A-7") {
			t.Errorf("warning missing problem description: %s", w)
		}
		if !strings.Contains(w, "pm link A-7 A-3 --type related") {
			t.Errorf("warning missing fix command: %s", w)
		}
	})

	t.Run("symmetric_related_no_duplicate_warning", func(t *testing.T) {
		// Both sides missing — should report only once (from A-3's perspective)
		rels := []relationshipData{
			{fromTicket: "A-3", toTicket: "A-7", relType: "related"},
		}
		warnings := validateRelationshipSymmetry(rels)
		if len(warnings) != 1 {
			t.Errorf("expected exactly 1 warning (no duplicate), got %d: %v", len(warnings), warnings)
		}
	})

	t.Run("empty_relationships_no_warnings", func(t *testing.T) {
		warnings := validateRelationshipSymmetry(nil)
		if len(warnings) != 0 {
			t.Errorf("expected no warnings for empty input, got %d", len(warnings))
		}
	})

	t.Run("fix_command_on_second_line", func(t *testing.T) {
		rels := []relationshipData{
			{fromTicket: "A-1", toTicket: "A-2", relType: "depends-on"},
		}
		warnings := validateRelationshipSymmetry(rels)
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(warnings))
		}
		lines := strings.Split(warnings[0], "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines in warning, got %d: %q", len(lines), warnings[0])
		}
		if !strings.HasPrefix(strings.TrimSpace(lines[1]), "pm link") {
			t.Errorf("second line should be the pm link command, got: %q", lines[1])
		}
	})
}
