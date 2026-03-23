package cache

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
)

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

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	ctx := context.Background()

	// Clear existing tickets using Bob
	deleteTickets := sqlite.Delete(dm.From("tickets"))
	deleteTicketsSQL, deleteTicketsArgs, err := deleteTickets.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build DELETE tickets query: %w", err)
	}
	_, err = tx.Exec(deleteTicketsSQL, deleteTicketsArgs...)
	if err != nil {
		return fmt.Errorf("failed to clear tickets table: %w", err)
	}

	// Clear existing comments using Bob
	deleteComments := sqlite.Delete(dm.From("comments"))
	deleteCommentsSQL, deleteCommentsArgs, err := deleteComments.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build DELETE comments query: %w", err)
	}
	_, err = tx.Exec(deleteCommentsSQL, deleteCommentsArgs...)
	if err != nil {
		return fmt.Errorf("failed to clear comments table: %w", err)
	}

	// Clear existing relationships using Bob
	deleteRelationships := sqlite.Delete(dm.From("relationships"))
	deleteRelationshipsSQL, deleteRelationshipsArgs, err := deleteRelationships.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build DELETE relationships query: %w", err)
	}
	_, err = tx.Exec(deleteRelationshipsSQL, deleteRelationshipsArgs...)
	if err != nil {
		return fmt.Errorf("failed to clear relationships table: %w", err)
	}

	// Scan tickets directory
	ticketsPath := filepath.Join(pmPath, "tickets")
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No tickets yet, just update timestamp
			return updateSyncTimestamp(tx)
		}
		return fmt.Errorf("failed to read tickets directory: %w", err)
	}

	// Collect all data for bulk inserts
	type ticketData struct {
		id        string
		title     string
		typ       string
		status    string
		priority  string
		assignee  string
		parent    string
		createdAt string
		updatedAt string
		body      string
		path      string
	}
	type commentData struct {
		ticketID  string
		author    string
		timestamp string
		filepath  string
	}
	type relationshipData struct {
		fromTicket string
		toTicket   string
		relType    string
	}

	var tickets []ticketData
	var comments []commentData
	var relationships []relationshipData

	// Process each ticket file
	for _, file := range files {
		// Handle both ticket files (.md) and comment directories
		if strings.HasSuffix(file.Name(), ".md") {
			// Process ticket file
			filePath := filepath.Join(ticketsPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue // Skip files we can't read
			}

			t, err := ticket.Parse(content)
			if err != nil {
				continue // Skip invalid tickets
			}

			// Extract body (everything after the second ---)
			parts := strings.SplitN(string(content), "---", 3)
			body := ""
			if len(parts) == 3 {
				body = strings.TrimSpace(parts[2])
			}

			tickets = append(tickets, ticketData{
				id:        t.ID,
				title:     t.Title,
				typ:       t.Type,
				status:    t.Status,
				priority:  t.Priority,
				assignee:  t.Assignee,
				parent:    t.Parent,
				createdAt: t.CreatedAt,
				updatedAt: t.UpdatedAt,
				body:      body,
			})

			// Collect relationships from depends_on array
			for _, depID := range t.DependsOn {
				relationships = append(relationships, relationshipData{
					fromTicket: t.ID,
					toTicket:   depID,
					relType:    "depends-on",
				})
			}

			// Collect relationships from blocks array
			for _, blockedID := range t.Blocks {
				relationships = append(relationships, relationshipData{
					fromTicket: t.ID,
					toTicket:   blockedID,
					relType:    "blocks",
				})
			}
		}
	}

	// Compute materialized paths for all tickets.
	// A materialized path encodes the full ancestor chain (e.g., "GPM-1/GPM-2/GPM-3"),
	// enabling subtree queries via a simple LIKE predicate instead of a recursive CTE.
	byID := make(map[string]*ticketData, len(tickets))
	for i := range tickets {
		byID[tickets[i].id] = &tickets[i]
	}

	var buildPath func(id string, visited map[string]bool) string
	buildPath = func(id string, visited map[string]bool) string {
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
		} else {
			t.path = buildPath(t.parent, visited) + "/" + id
		}
		return t.path
	}

	for i := range tickets {
		if tickets[i].path == "" {
			buildPath(tickets[i].id, make(map[string]bool))
		}
	}

	// Collect comments from comment directories
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		ticketID := file.Name()
		commentDir := filepath.Join(ticketsPath, ticketID)
		commentEntries, err := os.ReadDir(commentDir)
		if err != nil {
			continue // Skip if we can't read the directory
		}

		for _, commentFile := range commentEntries {
			if commentFile.IsDir() || !strings.HasSuffix(commentFile.Name(), ".md") {
				continue
			}

			commentPath := filepath.Join(commentDir, commentFile.Name())
			relPath := filepath.Join(ticketID, commentFile.Name())

			// Parse comment to get metadata
			comment, err := ticket.ParseCommentFile(commentPath)
			if err != nil {
				continue // Skip invalid comment files
			}

			comments = append(comments, commentData{
				ticketID:  ticketID,
				author:    comment.Author,
				timestamp: comment.CreatedAt.Format(time.RFC3339),
				filepath:  relPath,
			})
		}
	}

	// Bulk insert tickets using Bob
	if len(tickets) > 0 {
		insertTickets := sqlite.Insert(
			im.Into("tickets",
				"id", "title", "type", "status", "priority", "assignee", "parent", "created_at", "updated_at", "body", "path",
			),
		)
		
		for _, t := range tickets {
			insertTickets.Apply(
				im.Values(sqlite.Arg(t.id, t.title, t.typ, t.status, t.priority, t.assignee, t.parent, t.createdAt, t.updatedAt, t.body, t.path)),
			)
		}

		insertTicketsSQL, insertTicketsArgs, err := insertTickets.Build(ctx)
		if err != nil {
			return fmt.Errorf("failed to build INSERT tickets query: %w", err)
		}

		_, err = tx.Exec(insertTicketsSQL, insertTicketsArgs...)
		if err != nil {
			return fmt.Errorf("failed to insert tickets: %w", err)
		}
	}

	// Bulk insert comments using Bob
	if len(comments) > 0 {
		insertComments := sqlite.Insert(
			im.Into("comments", "ticket_id", "author", "timestamp", "filepath"),
		)

		for _, c := range comments {
			insertComments.Apply(
				im.Values(sqlite.Arg(c.ticketID, c.author, c.timestamp, c.filepath)),
			)
		}

		insertCommentsSQL, insertCommentsArgs, err := insertComments.Build(ctx)
		if err != nil {
			return fmt.Errorf("failed to build INSERT comments query: %w", err)
		}

		_, err = tx.Exec(insertCommentsSQL, insertCommentsArgs...)
		if err != nil {
			return fmt.Errorf("failed to insert comments: %w", err)
		}
	}

	// Bulk insert relationships using Bob
	if len(relationships) > 0 {
		insertRelationships := sqlite.Insert(
			im.Into("relationships", "from_ticket", "to_ticket", "relationship_type"),
		)

		for _, r := range relationships {
			insertRelationships.Apply(
				im.Values(sqlite.Arg(r.fromTicket, r.toTicket, r.relType)),
			)
		}

		insertRelationshipsSQL, insertRelationshipsArgs, err := insertRelationships.Build(ctx)
		if err != nil {
			return fmt.Errorf("failed to build INSERT relationships query: %w", err)
		}

		_, err = tx.Exec(insertRelationshipsSQL, insertRelationshipsArgs...)
		if err != nil {
			return fmt.Errorf("failed to insert relationships: %w", err)
		}
	}

	// Update sync timestamp
	if err := updateSyncTimestamp(tx); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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
