package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// setupListDB creates a temp-dir database with full schema and a set of fixture tickets.
// Ticket hierarchy:
//
//	ROOT-1 (status: todo)
//	  ├── CHILD-2 (status: in-progress, parent: ROOT-1)
//	  │     └── GRAND-3 (status: backlog, parent: CHILD-2)
//	  └── CHILD-4 (status: done, parent: ROOT-1)
//	ORPHAN-5 (status: todo, no parent)
func setupListDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	dir := t.TempDir()
	pmPath := filepath.Join(dir, ".pm")
	if err := os.MkdirAll(filepath.Join(pmPath, "tickets"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dbPath := filepath.Join(pmPath, ".cache.db")
	if err := RunMigrations(dbPath); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}

	fixtures := []struct {
		id, title, typ, status, parent, path string
	}{
		{"ROOT-1", "Root ticket", "story", "todo", "", "ROOT-1"},
		{"CHILD-2", "Child ticket", "task", "in-progress", "ROOT-1", "ROOT-1/CHILD-2"},
		{"GRAND-3", "Grandchild ticket", "task", "backlog", "CHILD-2", "ROOT-1/CHILD-2/GRAND-3"},
		{"CHILD-4", "Done child", "task", "done", "ROOT-1", "ROOT-1/CHILD-4"},
		{"ORPHAN-5", "Orphan ticket", "bug", "todo", "", "ORPHAN-5"},
	}
	for _, f := range fixtures {
		_, err := db.Exec(
			`INSERT INTO tickets (id, title, type, status, priority, assignee, parent, created_at, updated_at, body, path)
			 VALUES (?, ?, ?, ?, '', '', ?, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', '', ?)`,
			f.id, f.title, f.typ, f.status, f.parent, f.path,
		)
		if err != nil {
			t.Fatalf("insert %s: %v", f.id, err)
		}
	}
	return db, dbPath
}

// ids extracts the IDs from a result slice for easy assertion.
func ids(tickets []CachedTicket) []string {
	out := make([]string, len(tickets))
	for i, t := range tickets {
		out[i] = t.ID
	}
	return out
}

// containsID returns true if id is in the slice.
func containsID(tickets []CachedTicket, id string) bool {
	for _, t := range tickets {
		if t.ID == id {
			return true
		}
	}
	return false
}

func TestListTickets_TopLevel(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	tickets, err := ListTickets(db, ListOptions{})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}

	// Should return only top-level tickets: ROOT-1 and ORPHAN-5
	if !containsID(tickets, "ROOT-1") {
		t.Error("expected ROOT-1 in top-level results")
	}
	if !containsID(tickets, "ORPHAN-5") {
		t.Error("expected ORPHAN-5 in top-level results")
	}
	for _, id := range []string{"CHILD-2", "GRAND-3", "CHILD-4"} {
		if containsID(tickets, id) {
			t.Errorf("top-level should not include %s", id)
		}
	}
}

func TestListTickets_All(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	tickets, err := ListTickets(db, ListOptions{Subtree: true})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}

	if len(tickets) != 5 {
		t.Errorf("expected 5 tickets with --all, got %d: %v", len(tickets), ids(tickets))
	}
}

func TestListTickets_DirectChildren(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	tickets, err := ListTickets(db, ListOptions{ParentFilter: "ROOT-1"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}

	// Direct children of ROOT-1: CHILD-2, CHILD-4
	if !containsID(tickets, "CHILD-2") {
		t.Error("expected CHILD-2 as direct child of ROOT-1")
	}
	if !containsID(tickets, "CHILD-4") {
		t.Error("expected CHILD-4 as direct child of ROOT-1")
	}
	for _, id := range []string{"ROOT-1", "GRAND-3", "ORPHAN-5"} {
		if containsID(tickets, id) {
			t.Errorf("direct children should not include %s", id)
		}
	}
}

func TestListTickets_Subtree(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	tickets, err := ListTickets(db, ListOptions{ParentFilter: "ROOT-1", Subtree: true})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}

	// Subtree of ROOT-1: CHILD-2, GRAND-3, CHILD-4 (not ROOT-1 itself)
	for _, id := range []string{"CHILD-2", "GRAND-3", "CHILD-4"} {
		if !containsID(tickets, id) {
			t.Errorf("subtree should include %s", id)
		}
	}
	for _, id := range []string{"ROOT-1", "ORPHAN-5"} {
		if containsID(tickets, id) {
			t.Errorf("subtree should not include %s", id)
		}
	}
}

func TestListTickets_ParentNotFound(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	// Subtree mode: error when parent doesn't exist
	_, err := ListTickets(db, ListOptions{ParentFilter: "NONEXISTENT", Subtree: true})
	if err == nil {
		t.Error("expected error for nonexistent parent in subtree mode")
	}

	// Direct children mode: also an error when parent doesn't exist
	_, err = ListTickets(db, ListOptions{ParentFilter: "NONEXISTENT"})
	if err == nil {
		t.Error("expected error for nonexistent parent in direct-children mode")
	}
}

func TestListTickets_IncludeStates(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	// Include only "todo" tickets from all tickets
	tickets, err := ListTickets(db, ListOptions{
		Subtree:       true,
		IncludeStates: []string{"todo"},
	})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}

	for _, t2 := range tickets {
		if t2.Status != "todo" {
			t.Errorf("expected status=todo, got %q for %s", t2.Status, t2.ID)
		}
	}
	if !containsID(tickets, "ROOT-1") || !containsID(tickets, "ORPHAN-5") {
		t.Error("expected ROOT-1 and ORPHAN-5 in todo-filtered results")
	}
}

func TestListTickets_ExcludeStates(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	// Exclude "done" tickets from all tickets
	tickets, err := ListTickets(db, ListOptions{
		Subtree:       true,
		ExcludeStates: []string{"done"},
	})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}

	for _, t2 := range tickets {
		if t2.Status == "done" {
			t.Errorf("done ticket %s should be excluded", t2.ID)
		}
	}
	if containsID(tickets, "CHILD-4") {
		t.Error("CHILD-4 (done) should be excluded")
	}
}

func TestListTickets_HasChildren(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	tickets, err := ListTickets(db, ListOptions{Subtree: true})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}

	for _, t2 := range tickets {
		switch t2.ID {
		case "ROOT-1", "CHILD-2":
			if t2.HasChildren == 0 {
				t.Errorf("%s should have HasChildren > 0", t2.ID)
			}
		case "GRAND-3", "CHILD-4", "ORPHAN-5":
			if t2.HasChildren != 0 {
				t.Errorf("%s should have HasChildren == 0", t2.ID)
			}
		}
	}
}

func TestListTickets_CaseInsensitiveParent(t *testing.T) {
	db, _ := setupListDB(t)
	defer db.Close()

	// Pass lower-case parent ID — should still find CHILD-2 and CHILD-4
	tickets, err := ListTickets(db, ListOptions{ParentFilter: "root-1"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if !containsID(tickets, "CHILD-2") || !containsID(tickets, "CHILD-4") {
		t.Errorf("case-insensitive parent match failed, got: %v", ids(tickets))
	}
}

// setupRelationshipDB extends setupListDB with relationship fixtures:
//
//	CHILD-2  depends-on  ORPHAN-5   → (CHILD-2, ORPHAN-5, depends-on)
//	GRAND-3  depends-on  ROOT-1     → (GRAND-3, ROOT-1, depends-on)
//	CHILD-4  related     ORPHAN-5   → (CHILD-4, ORPHAN-5, related)
func setupRelationshipDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := setupListDB(t)

	rels := []struct{ from, to, typ string }{
		{"CHILD-2", "ORPHAN-5", "depends-on"},
		{"GRAND-3", "ROOT-1", "depends-on"},
		{"CHILD-4", "ORPHAN-5", "related"},
	}
	for _, r := range rels {
		_, err := db.Exec(
			`INSERT INTO relationships (from_ticket, to_ticket, relationship_type) VALUES (?, ?, ?)`,
			r.from, r.to, r.typ,
		)
		if err != nil {
			t.Fatalf("insert relationship %s->%s: %v", r.from, r.to, err)
		}
	}
	return db
}

func TestListTickets_DependsOn(t *testing.T) {
	db := setupRelationshipDB(t)
	defer db.Close()

	// Tickets that depend on ORPHAN-5: CHILD-2
	tickets, err := ListTickets(db, ListOptions{Subtree: true, DependsOn: "ORPHAN-5"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(tickets) != 1 || !containsID(tickets, "CHILD-2") {
		t.Errorf("expected [CHILD-2], got %v", ids(tickets))
	}

	// Tickets that depend on ROOT-1: GRAND-3
	tickets, err = ListTickets(db, ListOptions{Subtree: true, DependsOn: "ROOT-1"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(tickets) != 1 || !containsID(tickets, "GRAND-3") {
		t.Errorf("expected [GRAND-3], got %v", ids(tickets))
	}
}

func TestListTickets_DependsOn_CaseInsensitive(t *testing.T) {
	db := setupRelationshipDB(t)
	defer db.Close()

	tickets, err := ListTickets(db, ListOptions{Subtree: true, DependsOn: "orphan-5"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if !containsID(tickets, "CHILD-2") {
		t.Errorf("expected CHILD-2 with lowercase depends-on filter, got %v", ids(tickets))
	}
}

func TestListTickets_Blocks(t *testing.T) {
	db := setupRelationshipDB(t)
	defer db.Close()

	// Tickets blocking CHILD-2 (i.e., CHILD-2 depends on them): ORPHAN-5
	tickets, err := ListTickets(db, ListOptions{Subtree: true, Blocks: "CHILD-2"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(tickets) != 1 || !containsID(tickets, "ORPHAN-5") {
		t.Errorf("expected [ORPHAN-5], got %v", ids(tickets))
	}

	// Tickets blocking GRAND-3: ROOT-1
	tickets, err = ListTickets(db, ListOptions{Subtree: true, Blocks: "GRAND-3"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(tickets) != 1 || !containsID(tickets, "ROOT-1") {
		t.Errorf("expected [ROOT-1], got %v", ids(tickets))
	}
}

func TestListTickets_Related(t *testing.T) {
	db := setupRelationshipDB(t)
	defer db.Close()

	// ORPHAN-5 has: CHILD-2 depends-on it, CHILD-4 related to it
	tickets, err := ListTickets(db, ListOptions{Subtree: true, Related: "ORPHAN-5"})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if !containsID(tickets, "CHILD-2") {
		t.Errorf("expected CHILD-2 in related results, got %v", ids(tickets))
	}
	if !containsID(tickets, "CHILD-4") {
		t.Errorf("expected CHILD-4 in related results, got %v", ids(tickets))
	}
	// ORPHAN-5 itself should not appear
	if containsID(tickets, "ORPHAN-5") {
		t.Error("ORPHAN-5 should not appear in its own related results")
	}
}

func TestListTickets_DependsOn_CombinedWithStatus(t *testing.T) {
	db := setupRelationshipDB(t)
	defer db.Close()

	// CHILD-2 (in-progress) and nothing else depends on ORPHAN-5 with status todo
	tickets, err := ListTickets(db, ListOptions{
		Subtree:       true,
		DependsOn:     "ORPHAN-5",
		IncludeStates: []string{"in-progress"},
	})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(tickets) != 1 || !containsID(tickets, "CHILD-2") {
		t.Errorf("expected [CHILD-2], got %v", ids(tickets))
	}

	// Filter by a status CHILD-2 doesn't have — should return nothing
	tickets, err = ListTickets(db, ListOptions{
		Subtree:       true,
		DependsOn:     "ORPHAN-5",
		IncludeStates: []string{"done"},
	})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(tickets) != 0 {
		t.Errorf("expected no results, got %v", ids(tickets))
	}
}
