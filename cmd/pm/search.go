package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/ditsara/git-product-manager/internal/cache"
	"github.com/ditsara/git-product-manager/internal/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tickets by ID, title, or body content",
	Long: `Searches tickets for a query string across ID, title, and body content.

Results are ordered by relevance: ID matches first, then title, then body.
By default all statuses are included. Use status flags to narrow results.

Examples:
  pm search "authentication"
  pm search "login bug" --active
  pm search "refactor" --status in-progress
  pm search "closed work" --completed`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		pmPath := ".pm"

		statusFilter, _ := cmd.Flags().GetString("status")
		showAll, _ := cmd.Flags().GetBool("all")
		showCompleted, _ := cmd.Flags().GetBool("completed")
		showActive, _ := cmd.Flags().GetBool("active")
		showIncomplete, _ := cmd.Flags().GetBool("incomplete")

		if err := cache.EnsureCacheReady(pmPath); err != nil {
			log.Fatalf("Error initializing cache: %v", err)
		}

		shouldSync, err := cache.ShouldSync(pmPath)
		if err != nil {
			log.Printf("Warning: failed to check cache staleness: %v", err)
		} else if shouldSync {
			if err := cache.SyncCache(pmPath); err != nil {
				log.Fatalf("Error syncing cache: %v", err)
			}
		}

		dbPath := filepath.Join(pmPath, ".cache.db")
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer db.Close()

		workflowPath := filepath.Join(pmPath, "config", "workflow.yaml")
		workflow, err := config.LoadWorkflow(workflowPath)
		if err != nil {
			log.Fatalf("Error loading workflow: %v", err)
		}

		opts := cache.SearchOptions{}
		if statusFilter != "" {
			opts.IncludeStates = []string{statusFilter}
		} else if showCompleted {
			opts.IncludeStates = workflow.GetCompletedStates()
		} else if showActive {
			opts.IncludeStates = workflow.GetStateGroup("active")
		} else if showIncomplete {
			opts.IncludeStates = workflow.GetStateGroup("incomplete")
		} else if !showAll {
			// Default: all statuses (no filter)
		}

		results, err := cache.SearchTickets(db, query, opts)
		if err != nil {
			log.Fatalf("Error searching tickets: %v", err)
		}

		if len(results) == 0 {
			fmt.Printf("No results for %q\n", query)
			return
		}

		fmt.Printf("Search results for %q (%d match", query, len(results))
		if len(results) != 1 {
			fmt.Print("es")
		}
		fmt.Println("):")
		fmt.Println()

		for _, r := range results {
			fmt.Printf("%s: %s\n", r.ID, r.Title)
			fmt.Printf("  Type: %s | Status: %s\n", r.Type, r.Status)
			if r.Snippet != "" {
				fmt.Printf("  Match: %s\n", r.Snippet)
			}
			fmt.Println()
		}
	},
}

func init() {
	searchCmd.Flags().String("status", "", "Filter by specific status")
	searchCmd.Flags().Bool("all", false, "Show all statuses (default behaviour, explicit form)")
	searchCmd.Flags().Bool("completed", false, "Show only completed tickets")
	searchCmd.Flags().Bool("active", false, "Show only active tickets (todo, in-progress)")
	searchCmd.Flags().Bool("incomplete", false, "Show only incomplete tickets")

	searchCmd.RegisterFlagCompletionFunc("status", completeStates)

	rootCmd.AddCommand(searchCmd)
}
