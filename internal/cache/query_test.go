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
