package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
		milestoneFilter, _ := cmd.Flags().GetString("milestone")
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

		// Validate milestone existence (warn but continue)
		if milestoneFilter != "" {
			milestonePath := filepath.Join(pmPath, "milestones", milestoneFilter+".md")
			if _, err := os.Stat(milestonePath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: milestone '%s' not found\n", milestoneFilter)
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
		opts := cache.ListOptions{
			ParentFilter:    parentFilter,
			Subtree:         showAll || (milestoneFilter != "" && parentFilter == ""),
			MilestoneFilter: milestoneFilter,
		}

		if statusFilter != "" {
			opts.IncludeStates = []string{statusFilter}
		} else if showCompleted {
			opts.IncludeStates = workflow.GetCompletedStates()
		} else if showActive {
			opts.IncludeStates = workflow.GetStateGroup("active")
		} else if showIncomplete {
			opts.IncludeStates = workflow.GetStateGroup("incomplete")
		} else if !showAll {
			if completedStates := workflow.GetCompletedStates(); len(completedStates) > 0 {
				opts.ExcludeStates = completedStates
			}
		}

		tickets, err := cache.ListTickets(db, opts)
		if err != nil {
			log.Fatalf("Error querying tickets: %v", err)
		}

		listCols := []TableColumn{
			{Header: "ID", Width: 20},
			{Header: "TITLE", Width: 50},
			{Header: "TYPE", Width: 10},
			{Header: "STATUS", Width: 15},
		}
		var listRows [][]string
		for _, t := range tickets {
			displayID := t.ID
			if t.HasChildren > 0 {
				displayID = t.ID + " (+)"
			}
			listRows = append(listRows, []string{
				displayID,
				truncate(t.Title, 50),
				t.Type,
				t.Status,
			})
		}
		fmt.Println(renderTable(listCols, listRows, 3))

		if len(tickets) == 0 && parentFilter != "" {
			fmt.Printf("\nNo children found for %s\n", parentFilter)
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
	listCmd.Flags().String("milestone", "", "Filter tickets by milestone ID")

	// Register completion functions for flags
	listCmd.RegisterFlagCompletionFunc("status", completeStates)
	listCmd.RegisterFlagCompletionFunc("parent", completeTicketIDs)
	listCmd.RegisterFlagCompletionFunc("milestone", completeMilestoneIDs)

	rootCmd.AddCommand(listCmd)
}
