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
	Long:  `Lists all tickets from the .pm/tickets directory with optional filtering.`,
	Run: func(cmd *cobra.Command, cmdArgs []string) {
		statusFilter, _ := cmd.Flags().GetString("status")
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

		// Build query
		query := "SELECT id, title, type, status FROM tickets"
		var queryArgs []interface{}

		if statusFilter != "" {
			query += " WHERE status = ?"
			queryArgs = append(queryArgs, statusFilter)
		}

		query += " ORDER BY updated_at DESC"

		rows, err := db.Query(query, queryArgs...)
		if err != nil {
			log.Fatalf("Error querying tickets: %v", err)
		}
		defer rows.Close()

		fmt.Printf("%-20s %-50s %-10s %-15s\n", "ID", "TITLE", "TYPE", "STATUS")
		fmt.Println(strings.Repeat("-", 95))

		for rows.Next() {
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
	},
}

func init() {
	listCmd.Flags().String("status", "", "Filter by status")
	rootCmd.AddCommand(listCmd)
}
