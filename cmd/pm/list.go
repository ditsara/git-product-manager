package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/cache"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// truncate truncates a string to maxLen characters, adding "..." if truncated.
// Uses rune counting for proper Unicode/emoji handling.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return "..."
	}
	return string(runes[:maxLen-3]) + "..."
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tickets",
	Long: `Lists tickets from the .pm/tickets directory with optional filtering.

By default, shows only top-level tickets (those without a parent).

Examples:
  pm list                      # Show top-level tickets only
  pm list --all                # Show all tickets
  pm list --parent GPM-1       # Show direct children of GPM-1
  pm list --parent GPM-1 --all # Show entire subtree under GPM-1
  pm list --status todo        # Filter by status (works with all modes)`,
	Run: func(cmd *cobra.Command, cmdArgs []string) {
		statusFilter, _ := cmd.Flags().GetString("status")
		showAll, _ := cmd.Flags().GetBool("all")
		parentFilter, _ := cmd.Flags().GetString("parent")
		pmPath := ".pm"

		// Ensure cache database exists and has current schema
		if err := cache.EnsureCacheReady(pmPath); err != nil {
			log.Fatalf("Error initializing cache: %v", err)
		}

		// Check if cache needs sync and sync if necessary
		shouldSync, err := cache.ShouldSync(pmPath)
		if err != nil {
			log.Printf("Warning: failed to check cache staleness: %v", err)
			log.Println("Continuing with potentially stale cache...")
		} else if shouldSync {
			if err := cache.SyncCache(pmPath); err != nil {
				log.Fatalf("Error syncing cache: %v", err)
			}
		}

		// Query from SQLite cache
		dbPath := filepath.Join(pmPath, ".cache.db")
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer db.Close()

		// Validate parent ticket exists if specified
		var normalizedParent string
		if parentFilter != "" {
			foundTicketPath := findTicketByID(parentFilter)
			// Extract just the ticket ID from the path (strip .pm/tickets/ and .md)
			normalizedParent = strings.TrimSuffix(filepath.Base(foundTicketPath), ".md")
		}

		// Build query based on flags
		var query string
		var queryArgs []interface{}

		if parentFilter != "" && showAll {
			// Recursive subtree query
			query = `
				WITH RECURSIVE subtree(id) AS (
					SELECT id FROM tickets WHERE UPPER(parent) = UPPER(?)
					UNION ALL
					SELECT t.id FROM tickets t
					JOIN subtree s ON UPPER(t.parent) = UPPER(s.id)
				)
				SELECT id, title, type, status FROM tickets
				WHERE id IN (SELECT id FROM subtree)`
			queryArgs = append(queryArgs, normalizedParent)
		} else if parentFilter != "" {
			// Direct children only
			query = "SELECT id, title, type, status FROM tickets WHERE UPPER(parent) = UPPER(?)"
			queryArgs = append(queryArgs, normalizedParent)
		} else if !showAll {
			// Default: top-level tickets only (no parent)
			query = "SELECT id, title, type, status FROM tickets WHERE (parent IS NULL OR parent = '')"
		} else {
			// Show all tickets
			query = "SELECT id, title, type, status FROM tickets"
		}

		// Add status filter if specified
		if statusFilter != "" {
			if strings.Contains(query, "WHERE") {
				query += " AND status = ?"
			} else {
				query += " WHERE status = ?"
			}
			queryArgs = append(queryArgs, statusFilter)
		}

		query += " ORDER BY updated_at DESC"

		rows, err := db.Query(query, queryArgs...)
		if err != nil {
			log.Fatalf("Error querying tickets: %v", err)
		}
		defer rows.Close()

		// Check if we got any results
		hasResults := false
		fmt.Printf("%-20s %-50s %-10s %-15s\n", "ID", "TITLE", "TYPE", "STATUS")
		fmt.Println(strings.Repeat("-", 95))

		for rows.Next() {
			hasResults = true
			var id, title, ticketType, status string
			if err := rows.Scan(&id, &title, &ticketType, &status); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}

			fmt.Printf("%-20s %-50s %-10s %-15s\n", id, truncate(title, 50), ticketType, status)
		}

		if err := rows.Err(); err != nil {
			log.Fatalf("Error iterating rows: %v", err)
		}

		// Show helpful message if parent filter returned no results
		if !hasResults && parentFilter != "" {
			fmt.Printf("\nNo children found for %s\n", normalizedParent)
		}
	},
}

func init() {
	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().Bool("all", false, "Show all tickets (default: top-level only)")
	listCmd.Flags().String("parent", "", "Show children of specified ticket ID")
	rootCmd.AddCommand(listCmd)
}
