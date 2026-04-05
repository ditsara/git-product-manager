package cache

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/milestone"
	"github.com/ditsara/git-product-manager/internal/ticket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
)

// ticketData holds the fields extracted from a ticket file for bulk cache insertion.
type ticketData struct {
	id         string
	title      string
	typ        string
	status     string
	priority   string
	assignee   string
	parent     string
	createdAt  string
	updatedAt  string
	body       string
	path       string
	milestones string
}

// commentData holds comment metadata for bulk cache insertion.
type commentData struct {
	ticketID  string
	author    string
	timestamp string
	filepath  string
}

// relationshipData holds a single directed relationship for bulk cache insertion.
type relationshipData struct {
	fromTicket string
	toTicket   string
	relType    string
}

// clearTable deletes all rows from the named table within the given transaction.
func clearTable(ctx context.Context, tx *sql.Tx, table string) error {
	q := sqlite.Delete(dm.From(table))
	sql, args, err := q.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build DELETE %s query: %w", table, err)
	}
	if _, err = tx.Exec(sql, args...); err != nil {
		return fmt.Errorf("failed to clear %s table: %w", table, err)
	}
	return nil
}

// buildPath computes the materialized path for a ticket: the full ancestor chain
// separated by slashes (e.g. "GPM-1/GPM-2/GPM-3"). Orphans and cycle members
// fall back to the bare ticket ID. Computed paths are memoised in the struct.
func buildPath(id string, byID map[string]*ticketData, visited map[string]bool) string {
	t, ok := byID[id]
	if !ok {
		return id
	}
	if t.path != "" {
		return t.path // already computed
	}
	if visited[id] {
		t.path = id // cycle detected — fall back to bare ID
		return t.path
	}
	visited[id] = true
	if t.parent == "" {
		t.path = id
	} else if _, parentExists := byID[t.parent]; !parentExists {
		t.path = id // orphan: parent not in ticket set, fall back to bare ID
	} else {
		t.path = buildPath(t.parent, byID, visited) + "/" + id
	}
	return t.path
}

// ShouldSync checks if the cache needs to be synchronized with the filesystem
// by comparing the last sync timestamp with the modification times of ticket files
func ShouldSync(pmPath string) (bool, error) {
	dbPath := filepath.Join(pmPath, ".cache.db")

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get last sync timestamp
	var lastSyncStr string
	err = db.QueryRow("SELECT value FROM cache_metadata WHERE key = 'last_sync_timestamp'").Scan(&lastSyncStr)
	if err != nil {
		// If metadata doesn't exist, we need to sync
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, fmt.Errorf("failed to get last sync timestamp: %w", err)
	}

	lastSync, err := time.Parse(time.RFC3339, lastSyncStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse last sync timestamp: %w", err)
	}

	// Check if any ticket file is newer than last sync
	ticketsPath := filepath.Join(pmPath, "tickets")
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		// If tickets directory doesn't exist, no need to sync
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read tickets directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// TIMING NOTE: Use truncated comparison to account for filesystem timestamp precision.
		// Most filesystems store mtime with second-level precision, but Go's time.Time uses
		// nanosecond precision. Truncating both times to seconds ensures we don't get false
		// positives from subsecond differences between when we record the sync time and when
		// the filesystem actually stored the file's mtime.
		//
		// We use !Before (>=) instead of After (>) to catch the edge case where a file is
		// modified in the same second as the sync timestamp. This ensures files changed during
		// rapid operations (tests, scripts) are properly detected and re-synced.
		fileTime := info.ModTime().Truncate(time.Second)
		syncTime := lastSync.Truncate(time.Second)

		if !fileTime.Before(syncTime) {
			return true, nil
		}
	}

	return false, nil
}

// SyncCache synchronizes the SQLite cache with the filesystem
// by scanning all ticket files and updating the database
func SyncCache(pmPath string) error {
	dbPath := filepath.Join(pmPath, ".cache.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	ctx := context.Background()

	for _, table := range []string{"tickets", "comments", "relationships"} {
		if err := clearTable(ctx, tx, table); err != nil {
			return err
		}
	}

	ticketsPath := filepath.Join(pmPath, "tickets")
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return updateSyncTimestamp(tx)
		}
		return fmt.Errorf("failed to read tickets directory: %w", err)
	}

	tickets, relationships := scanTicketFiles(ticketsPath, files)
	comments := scanCommentDirs(ticketsPath, files)

	if err := syncTickets(ctx, tx, tickets); err != nil {
		return err
	}
	if err := syncComments(ctx, tx, comments); err != nil {
		return err
	}
	if err := syncRelationships(ctx, tx, relationships); err != nil {
		return err
	}
	if err := updateSyncTimestamp(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Sync milestones independently after the ticket transaction commits
	if err := SyncMilestones(pmPath); err != nil {
		// Non-fatal: milestones dir may not exist in older projects
		_ = err
	}

	return nil
}

// scanTicketFiles reads all .md ticket files from ticketsPath, parsing each into
// ticketData and collecting any relationship edges declared in the ticket front matter.
// Materialized paths are computed across all tickets after parsing.
func scanTicketFiles(ticketsPath string, files []os.DirEntry) ([]ticketData, []relationshipData) {
	var tickets []ticketData
	var relationships []relationshipData

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		filePath := filepath.Join(ticketsPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		t, err := ticket.Parse(content)
		if err != nil {
			continue
		}
		parts := strings.SplitN(string(content), "---", 3)
		body := ""
		if len(parts) == 3 {
			body = strings.TrimSpace(parts[2])
		}
		tickets = append(tickets, ticketData{
			id:         t.ID,
			title:      t.Title,
			typ:        t.Type,
			status:     t.Status,
			priority:   t.Priority,
			assignee:   t.Assignee,
			parent:     t.Parent,
			createdAt:  t.CreatedAt,
			updatedAt:  t.UpdatedAt,
			body:       body,
			milestones: strings.Join(t.Milestones, ","),
		})
		for _, depID := range t.DependsOn {
			relationships = append(relationships, relationshipData{fromTicket: t.ID, toTicket: depID, relType: "depends-on"})
		}
		for _, blockedID := range t.Blocks {
			relationships = append(relationships, relationshipData{fromTicket: t.ID, toTicket: blockedID, relType: "blocks"})
		}
	}

	// Compute materialized paths for all tickets.
	// A materialized path encodes the full ancestor chain (e.g., "GPM-1/GPM-2/GPM-3"),
	// enabling subtree queries via a simple LIKE predicate instead of a recursive CTE.
	byID := make(map[string]*ticketData, len(tickets))
	for i := range tickets {
		byID[tickets[i].id] = &tickets[i]
	}
	for i := range tickets {
		if tickets[i].path == "" {
			buildPath(tickets[i].id, byID, make(map[string]bool))
		}
	}

	return tickets, relationships
}

// scanCommentDirs reads all comment directories from ticketsPath and returns
// a flat list of commentData for bulk insertion.
func scanCommentDirs(ticketsPath string, files []os.DirEntry) []commentData {
	var comments []commentData

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		ticketID := file.Name()
		commentDir := filepath.Join(ticketsPath, ticketID)
		commentEntries, err := os.ReadDir(commentDir)
		if err != nil {
			continue
		}
		for _, commentFile := range commentEntries {
			if commentFile.IsDir() || !strings.HasSuffix(commentFile.Name(), ".md") {
				continue
			}
			commentPath := filepath.Join(commentDir, commentFile.Name())
			relPath := filepath.Join(ticketID, commentFile.Name())
			c, err := ticket.ParseCommentFile(commentPath)
			if err != nil {
				continue
			}
			comments = append(comments, commentData{
				ticketID:  ticketID,
				author:    c.Author,
				timestamp: c.CreatedAt.Format(time.RFC3339),
				filepath:  relPath,
			})
		}
	}

	return comments
}

// updateSyncTimestamp updates the last_sync_timestamp in cache_metadata
func updateSyncTimestamp(tx *sql.Tx) error {
	// Record current time. Note: We use !Before (>=) comparison in ShouldSync,
	// which means files with mtime >= sync_time will trigger a sync.
	// This is intentional to catch files modified in the same second as the sync.
	now := time.Now().UTC().Format(time.RFC3339)
	
	ctx := context.Background()
	
	// Use Bob for INSERT OR REPLACE
	updateQuery := sqlite.Insert(
		im.Into("cache_metadata", "key", "value"),
		im.Values(sqlite.Arg("last_sync_timestamp", now)),
		im.OrReplace(),
	)
	
	updateSQL, updateArgs, err := updateQuery.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build UPDATE cache_metadata query: %w", err)
	}
	
	_, err = tx.Exec(updateSQL, updateArgs...)
	if err != nil {
		return fmt.Errorf("failed to update sync timestamp: %w", err)
	}
	return nil
}

// syncTickets bulk-inserts ticket rows into the cache within an existing transaction.
func syncTickets(ctx context.Context, tx *sql.Tx, tickets []ticketData) error {
	if len(tickets) == 0 {
		return nil
	}
	q := sqlite.Insert(
		im.Into("tickets",
			"id", "title", "type", "status", "priority", "assignee", "parent", "created_at", "updated_at", "body", "path", "milestones",
		),
	)
	for _, t := range tickets {
		q.Apply(im.Values(sqlite.Arg(t.id, t.title, t.typ, t.status, t.priority, t.assignee, t.parent, t.createdAt, t.updatedAt, t.body, t.path, t.milestones)))
	}
	querySQL, queryArgs, err := q.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build INSERT tickets query: %w", err)
	}
	if _, err = tx.Exec(querySQL, queryArgs...); err != nil {
		return fmt.Errorf("failed to insert tickets: %w", err)
	}
	return nil
}

// syncComments bulk-inserts comment rows into the cache within an existing transaction.
// OR IGNORE handles the edge case where two comments are created in the same second by
// the same author on the same ticket (identical PRIMARY KEY). Files on disk are the
// source of truth; the cache keeps the first one.
func syncComments(ctx context.Context, tx *sql.Tx, comments []commentData) error {
	if len(comments) == 0 {
		return nil
	}
	q := sqlite.Insert(
		im.Into("comments", "ticket_id", "author", "timestamp", "filepath"),
		im.OrIgnore(),
	)
	for _, c := range comments {
		q.Apply(im.Values(sqlite.Arg(c.ticketID, c.author, c.timestamp, c.filepath)))
	}
	querySQL, queryArgs, err := q.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build INSERT comments query: %w", err)
	}
	if _, err = tx.Exec(querySQL, queryArgs...); err != nil {
		return fmt.Errorf("failed to insert comments: %w", err)
	}
	return nil
}

// syncRelationships bulk-inserts relationship rows into the cache within an existing transaction.
func syncRelationships(ctx context.Context, tx *sql.Tx, relationships []relationshipData) error {
	if len(relationships) == 0 {
		return nil
	}
	q := sqlite.Insert(
		im.Into("relationships", "from_ticket", "to_ticket", "relationship_type"),
	)
	for _, r := range relationships {
		q.Apply(im.Values(sqlite.Arg(r.fromTicket, r.toTicket, r.relType)))
	}
	querySQL, queryArgs, err := q.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build INSERT relationships query: %w", err)
	}
	if _, err = tx.Exec(querySQL, queryArgs...); err != nil {
		return fmt.Errorf("failed to insert relationships: %w", err)
	}
	return nil
}

// SyncMilestones reads all milestone files from .pm/milestones/ and upserts
// them into the milestones cache table.
func SyncMilestones(pmPath string) error {
	dbPath := filepath.Join(pmPath, ".cache.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing milestones
	if _, err := tx.Exec("DELETE FROM milestones"); err != nil {
		return fmt.Errorf("failed to clear milestones table: %w", err)
	}

	milestonesPath := filepath.Join(pmPath, "milestones")
	milestones, err := milestone.ListMilestones(milestonesPath)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			// No milestones directory yet — nothing to sync
			return tx.Commit()
		}
		return fmt.Errorf("failed to list milestones: %w", err)
	}

	for _, m := range milestones {
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO milestones (id, title, description, due_date, state, created_at, closed_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			m.ID, m.Title, m.Description, m.DueDate, m.State, m.CreatedAt, m.ClosedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert milestone %s: %w", m.ID, err)
		}
	}

	return tx.Commit()
}
