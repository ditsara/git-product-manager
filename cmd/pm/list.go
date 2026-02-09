package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/cache"
	"github.com/ditsara/git-product-manager/internal/config"
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
	Short: "List tickets",
	Aliases: []string{"ls"},
	Long: `Lists tickets from the .pm/tickets directory with optional filtering.

By default, shows only top-level incomplete tickets (hides completed work).
Use --all to show everything, or --completed to see only finished tickets.

Examples:
  pm list                      # Top-level incomplete tickets (default)
  pm list --all                # All tickets (including completed)
  pm list --completed          # Only completed tickets
  pm list --active             # Only active work (todo, in-progress)
  pm list --parent GPM-1       # Show direct children of GPM-1
  pm list --status todo        # Filter by specific status`,
	Run: func(cmd *cobra.Command, cmdArgs []string) {
		statusFilter, _ := cmd.Flags().GetString("status")
		showAll, _ := cmd.Flags().GetBool("all")
		showCompleted, _ := cmd.Flags().GetBool("completed")
		showActive, _ := cmd.Flags().GetBool("active")
		showIncomplete, _ := cmd.Flags().GetBool("incomplete")
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

		// Load workflow to get state_groups for filtering
		workflowPath := filepath.Join(pmPath, "config", "workflow.yaml")
		workflow, err := config.LoadWorkflow(workflowPath)
		if err != nil {
			log.Fatalf("Error loading workflow: %v", err)
		}

		// Determine which states to filter based on flags
		var includeStates []string
		var excludeStates []string

		if statusFilter != "" {
			// Explicit status filter takes precedence
			includeStates = []string{statusFilter}
		} else if showCompleted {
			// Show only completed states
			includeStates = workflow.GetCompletedStates()
		} else if showActive {
			// Show only active states
			includeStates = workflow.GetStateGroup("active")
		} else if showIncomplete {
			// Show only incomplete states
			includeStates = workflow.GetStateGroup("incomplete")
		} else if !showAll {
			// Default: exclude completed states (if defined)
			if completedStates := workflow.GetCompletedStates(); len(completedStates) > 0 {
				excludeStates = completedStates
			}
		}
		// If showAll, no filtering on states

		// Validate parent ticket exists if specified
		var normalizedParent string
		if parentFilter != "" {
			foundTicketPath := getTicketPath(parentFilter)
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
				SELECT id, title, type, status,
					CASE WHEN EXISTS(SELECT 1 FROM tickets AS t WHERE UPPER(t.parent) = UPPER(tickets.id))
						THEN 1 ELSE 0 END AS has_children
				FROM tickets
				WHERE id IN (SELECT id FROM subtree)`
			queryArgs = append(queryArgs, normalizedParent)
		} else if parentFilter != "" {
			// Direct children only
			query = `
				SELECT id, title, type, status,
					CASE WHEN EXISTS(SELECT 1 FROM tickets AS t WHERE UPPER(t.parent) = UPPER(tickets.id))
						THEN 1 ELSE 0 END AS has_children
				FROM tickets WHERE UPPER(parent) = UPPER(?)`
			queryArgs = append(queryArgs, normalizedParent)
		} else if !showAll {
			// Default: top-level tickets only (no parent)
			query = `
				SELECT id, title, type, status,
					CASE WHEN EXISTS(SELECT 1 FROM tickets AS t WHERE UPPER(t.parent) = UPPER(tickets.id))
						THEN 1 ELSE 0 END AS has_children
				FROM tickets WHERE (parent IS NULL OR parent = '')`
		} else {
			// Show all tickets
			query = `
				SELECT id, title, type, status,
					CASE WHEN EXISTS(SELECT 1 FROM tickets AS t WHERE UPPER(t.parent) = UPPER(tickets.id))
						THEN 1 ELSE 0 END AS has_children
				FROM tickets`
		}

		// Add status filters
		if len(includeStates) > 0 {
			// Include only these states
			placeholders := strings.Repeat("?,", len(includeStates))
			placeholders = placeholders[:len(placeholders)-1]

			if strings.Contains(query, "WHERE") {
				query += " AND status IN (" + placeholders + ")"
			} else {
				query += " WHERE status IN (" + placeholders + ")"
			}

			for _, state := range includeStates {
				queryArgs = append(queryArgs, state)
			}
		} else if len(excludeStates) > 0 {
			// Exclude these states
			placeholders := strings.Repeat("?,", len(excludeStates))
			placeholders = placeholders[:len(placeholders)-1]

			if strings.Contains(query, "WHERE") {
				query += " AND status NOT IN (" + placeholders + ")"
			} else {
				query += " WHERE status NOT IN (" + placeholders + ")"
			}

			for _, state := range excludeStates {
				queryArgs = append(queryArgs, state)
			}
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
			var hasChildren int
			if err := rows.Scan(&id, &title, &ticketType, &status, &hasChildren); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}

			// Append " (+)" to ID if ticket has children
			displayID := id
			if hasChildren > 0 {
				displayID = id + " (+)"
			}

			fmt.Printf("%-20s %-50s %-10s %-15s\n",
				displayID, truncate(title, 50), ticketType, status)
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
	listCmd.Flags().String("status", "", "Filter by specific status")
	listCmd.Flags().Bool("all", false, "Show all tickets (including completed)")
	listCmd.Flags().Bool("completed", false, "Show only completed tickets")
	listCmd.Flags().Bool("active", false, "Show only active tickets (todo, in-progress)")
	listCmd.Flags().Bool("incomplete", false, "Show only incomplete tickets")
	listCmd.Flags().StringP("parent", "P", "", "Show direct children of specified ticket ID")

	// Register completion functions for flags
	listCmd.RegisterFlagCompletionFunc("status", completeStates)
	listCmd.RegisterFlagCompletionFunc("parent", completeTicketIDs)

	rootCmd.AddCommand(listCmd)
}
