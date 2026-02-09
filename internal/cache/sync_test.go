package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShouldSync(t *testing.T) {
	// Create test workspace
	tempDir := t.TempDir()
	pmPath := filepath.Join(tempDir, ".pm")
	ticketsPath := filepath.Join(pmPath, "tickets")
	
	if err := os.MkdirAll(ticketsPath, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Initialize database with embedded migrations
	dbPath := filepath.Join(pmPath, ".cache.db")
	
	if err := RunMigrations(dbPath); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Test 1: Fresh database should need sync
	t.Run("fresh_database", func(t *testing.T) {
		shouldSync, err := ShouldSync(pmPath)
		if err != nil {
			t.Fatalf("ShouldSync() error = %v", err)
		}
		if shouldSync {
			t.Error("ShouldSync() = true, want false (no tickets yet)")
		}
	})

	// Test 2: Create a ticket file and check if sync is needed
	t.Run("new_ticket_file", func(t *testing.T) {
		ticketContent := `---
id: TEST-1
title: Test Ticket
type: task
status: todo
created_at: 2026-01-31T09:00:00Z
updated_at: 2026-01-31T09:00:00Z
---
Test content
`
		ticketPath := filepath.Join(ticketsPath, "TEST-1.md")
		if err := os.WriteFile(ticketPath, []byte(ticketContent), 0644); err != nil {
			t.Fatalf("Failed to create ticket file: %v", err)
		}

		// TIMING NOTE: Small delay ensures the file's mtime is detectably different from
		// the initial database timestamp. Without this, the test may pass or fail randomly
		// depending on how fast the test executes.
		time.Sleep(100 * time.Millisecond)

		shouldSync, err := ShouldSync(pmPath)
		if err != nil {
			t.Fatalf("ShouldSync() error = %v", err)
		}
		if !shouldSync {
			t.Error("ShouldSync() = false, want true (new ticket file)")
		}
	})

	// Test 3: After sync, should not need sync again
	t.Run("after_sync", func(t *testing.T) {
		if err := SyncCache(pmPath); err != nil {
			t.Fatalf("SyncCache() error = %v", err)
		}

		shouldSync, err := ShouldSync(pmPath)
		if err != nil {
			t.Fatalf("ShouldSync() error = %v", err)
		}
		if shouldSync {
			t.Error("ShouldSync() = true, want false (just synced)")
		}
	})
}

func TestSyncCache(t *testing.T) {
	// Create test workspace
	tempDir := t.TempDir()
	pmPath := filepath.Join(tempDir, ".pm")
	ticketsPath := filepath.Join(pmPath, "tickets")
	
	if err := os.MkdirAll(ticketsPath, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Initialize database with embedded migrations
	dbPath := filepath.Join(pmPath, ".cache.db")
	
	if err := RunMigrations(dbPath); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test tickets
	tickets := []struct {
		id      string
		title   string
		content string
	}{
		{
			id:    "TEST-1",
			title: "First Ticket",
			content: `---
id: TEST-1
title: First Ticket
type: task
status: todo
priority: high
assignee: alice
parent: ""
created_at: 2026-01-31T09:00:00Z
updated_at: 2026-01-31T09:00:00Z
---
# Description
First ticket description
`,
		},
		{
			id:    "TEST-2",
			title: "Second Ticket",
			content: `---
id: TEST-2
title: Second Ticket
type: bug
status: in-progress
priority: medium
assignee: bob
parent: ""
created_at: 2026-01-31T10:00:00Z
updated_at: 2026-01-31T10:00:00Z
---
# Bug Report
Second ticket description
`,
		},
	}

	for _, ticket := range tickets {
		ticketPath := filepath.Join(ticketsPath, ticket.id+".md")
		if err := os.WriteFile(ticketPath, []byte(ticket.content), 0644); err != nil {
			t.Fatalf("Failed to create ticket file %s: %v", ticket.id, err)
		}
	}

	// Test sync
	t.Run("sync_multiple_tickets", func(t *testing.T) {
		if err := SyncCache(pmPath); err != nil {
			t.Fatalf("SyncCache() error = %v", err)
		}

		// Verify tickets are in database
		db, err := openDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		rows, err := db.Query("SELECT id, title, type, status FROM tickets ORDER BY id")
		if err != nil {
			t.Fatalf("Failed to query tickets: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var id, title, ticketType, status string
			if err := rows.Scan(&id, &title, &ticketType, &status); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}

			if count < len(tickets) {
				if id != tickets[count].id {
					t.Errorf("Ticket %d: id = %s, want %s", count, id, tickets[count].id)
				}
				if title != tickets[count].title {
					t.Errorf("Ticket %d: title = %s, want %s", count, title, tickets[count].title)
				}
			}
			count++
		}

		if count != len(tickets) {
			t.Errorf("SyncCache() synced %d tickets, want %d", count, len(tickets))
		}
	})

	// Test re-sync (should replace existing data)
	t.Run("resync_updates_data", func(t *testing.T) {
		// Modify a ticket
		updatedContent := `---
id: TEST-1
title: Updated First Ticket
type: task
status: done
priority: high
assignee: alice
parent: ""
created_at: 2026-01-31T09:00:00Z
updated_at: 2026-02-01T09:00:00Z
---
# Description
Updated description
`
		ticketPath := filepath.Join(ticketsPath, "TEST-1.md")
		if err := os.WriteFile(ticketPath, []byte(updatedContent), 0644); err != nil {
			t.Fatalf("Failed to update ticket file: %v", err)
		}

		if err := SyncCache(pmPath); err != nil {
			t.Fatalf("SyncCache() error = %v", err)
		}

		// Verify update
		db, err := openDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		var title, status string
		err = db.QueryRow("SELECT title, status FROM tickets WHERE id = 'TEST-1'").Scan(&title, &status)
		if err != nil {
			t.Fatalf("Failed to query updated ticket: %v", err)
		}

		if title != "Updated First Ticket" {
			t.Errorf("Updated ticket title = %s, want 'Updated First Ticket'", title)
		}
			if status != "done" {
				t.Errorf("Updated ticket status = %s, want 'done'", status)
			}
		})
}
