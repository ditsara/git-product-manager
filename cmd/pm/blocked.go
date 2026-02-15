package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/cache"
	"github.com/ditsara/git-product-manager/internal/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

var blockedCmd = &cobra.Command{
	Use:   "blocked [ticket-id]",
	Short: "Show dependency and blocking relationships",
	Long: `Show dependency and blocking relationships for tickets.

Without a ticket ID, lists all tickets that have unresolved dependencies.
With a ticket ID, shows what blocks the ticket and what it blocks.

Examples:
  pm blocked              # List all blocked tickets
  pm blocked GPM-47       # Show dependencies for GPM-47`,
	ValidArgsFunction: completeTicketIDs,
	Run: func(cmd *cobra.Command, args []string) {
		pmPath := ".pm"

		// Ensure cache database exists and has current schema
		if err := cache.EnsureCacheReady(pmPath); err != nil {
			log.Fatalf("Error initializing cache: %v", err)
		}

		// Check if cache needs sync and sync if necessary
		shouldSync, err := cache.ShouldSync(pmPath)
		if err != nil {
			log.Fatalf("Error checking cache sync status: %v", err)
		}

		if shouldSync {
			if err := cache.SyncCache(pmPath); err != nil {
				log.Fatalf("Error syncing cache: %v", err)
			}
		}

		// Load workflow config to get completed states
		workflowPath := filepath.Join(pmPath, "config", "workflow.yaml")
		workflow, err := config.LoadWorkflow(workflowPath)
		if err != nil {
			log.Fatalf("Error loading workflow config: %v", err)
		}

		completedStates := workflow.GetCompletedStates()
		if len(completedStates) == 0 {
			log.Fatalf("No completed states defined in workflow.yaml")
		}

		// Open database
		dbPath := filepath.Join(pmPath, ".cache.db")
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer db.Close()

		if len(args) == 0 {
			// Global view: show all tickets with unresolved dependencies
			showGlobalBlockedView(db, completedStates)
		} else {
			// Specific ticket view
			ticketID := strings.ToUpper(args[0])
			showTicketBlockedView(db, ticketID, workflow)
		}
	},
}

// showGlobalBlockedView lists all tickets that have unresolved dependencies
func showGlobalBlockedView(db *sql.DB, completedStates []string) {
	// This query uses GROUP_CONCAT and complex HAVING with dynamic NOT IN clause
	// Bob doesn't have clean support for SQLite's GROUP_CONCAT or dynamic IN clauses in HAVING
	// Keeping as raw SQL per GPM-64 acceptance criteria: "May need to fallback to raw SQL for parts if Bob can't express it"
	
	// Build the NOT IN clause for completed states
	placeholders := make([]string, len(completedStates))
	args := make([]interface{}, len(completedStates))
	for i, state := range completedStates {
		placeholders[i] = "?"
		args[i] = state
	}
	notInClause := strings.Join(placeholders, ", ")

	query := fmt.Sprintf(`
		SELECT DISTINCT t.id, t.title, t.status,
		  GROUP_CONCAT(r.to_ticket || ':' || dep.title || ':' || dep.status, '|||') as blockers
		FROM tickets t
		JOIN relationships r ON r.from_ticket = t.id AND r.relationship_type = 'depends-on'
		JOIN tickets dep ON dep.id = r.to_ticket
		GROUP BY t.id, t.title, t.status
		HAVING COUNT(CASE WHEN dep.status NOT IN (%s) THEN 1 END) > 0
		ORDER BY t.updated_at DESC
	`, notInClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalf("Error querying blocked tickets: %v", err)
	}
	defer rows.Close()

	type blockedTicket struct {
		ID       string
		Title    string
		Status   string
		Blockers []struct {
			ID     string
			Title  string
			Status string
		}
	}

	var tickets []blockedTicket
	uniqueDeps := make(map[string]bool)

	for rows.Next() {
		var id, title, status, blockersStr string
		if err := rows.Scan(&id, &title, &status, &blockersStr); err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		ticket := blockedTicket{
			ID:     id,
			Title:  title,
			Status: status,
		}

		// Parse blockers string
		if blockersStr != "" {
			blockerParts := strings.Split(blockersStr, "|||")
			for _, bp := range blockerParts {
				parts := strings.Split(bp, ":")
				if len(parts) == 3 {
					ticket.Blockers = append(ticket.Blockers, struct {
						ID     string
						Title  string
						Status string
					}{
						ID:     parts[0],
						Title:  parts[1],
						Status: parts[2],
					})
					uniqueDeps[parts[0]] = true
				}
			}
		}

		tickets = append(tickets, ticket)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	// Display results
	if len(tickets) == 0 {
		fmt.Println("No blocked tickets found")
		return
	}

	fmt.Println("Blocked Tickets:")
	fmt.Println()

	for _, ticket := range tickets {
		fmt.Printf("%s: %s\n", ticket.ID, truncate(ticket.Title, 60))
		for _, blocker := range ticket.Blockers {
			fmt.Printf("  Blocked by: %s (%s)\n", blocker.ID, truncate(blocker.Title, 50))
		}
		fmt.Printf("  Status: %s\n", ticket.Status)
		fmt.Println()
	}

	fmt.Printf("%d ticket(s) blocked by %d dependenc(ies)\n", len(tickets), len(uniqueDeps))
}

// showTicketBlockedView shows dependencies and blockers for a specific ticket
func showTicketBlockedView(db *sql.DB, ticketID string, workflow *config.Workflow) {
	// Verify ticket exists using Bob query builder
	query := sqlite.Select(
		sm.Columns("title", "status"),
		sm.From("tickets"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(ticketID))),
	)

	// Build the SQL query
	querySQL, args, err := query.Build(context.Background())
	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	var title, status string
	err = db.QueryRow(querySQL, args...).Scan(&title, &status)
	if err == sql.ErrNoRows {
		log.Fatalf("Ticket not found: %s", ticketID)
	} else if err != nil {
		log.Fatalf("Error querying ticket: %v", err)
	}

	fmt.Printf("%s: %s\n\n", ticketID, truncate(title, 60))

	// Query what this ticket depends on
	// Note: Keeping as raw SQL - see GPM-67 for Bob migration
	dependsOnQuery := `
		SELECT r.to_ticket, t.title, t.status
		FROM relationships r
		JOIN tickets t ON t.id = r.to_ticket
		WHERE r.from_ticket = ? AND r.relationship_type = 'depends-on'
		ORDER BY r.to_ticket
	`

	rows, err := db.Query(dependsOnQuery, ticketID)
	if err != nil {
		log.Fatalf("Error querying dependencies: %v", err)
	}

	hasDeps := false
	fmt.Println("This ticket depends on:")
	for rows.Next() {
		var depID, depTitle, depStatus string
		if err := rows.Scan(&depID, &depTitle, &depStatus); err != nil {
			log.Fatalf("Error scanning dependency: %v", err)
		}
		hasDeps = true

		// Check if resolved
		indicator := "✗"
		if workflow.IsCompleted(depStatus) {
			indicator = "✓"
		}

		fmt.Printf("  • %s: %s (%s) %s\n", depID, truncate(depTitle, 40), depStatus, indicator)
	}
	rows.Close()

	if !hasDeps {
		fmt.Println("  (none)")
	}
	fmt.Println()

	// Query what depends on this ticket (reverse lookup)
	// Note: Keeping as raw SQL - see GPM-67 for Bob migration
	blocksQuery := `
		SELECT r.from_ticket, t.title, t.status
		FROM relationships r
		JOIN tickets t ON t.id = r.from_ticket
		WHERE r.to_ticket = ? AND r.relationship_type = 'depends-on'
		ORDER BY r.from_ticket
	`

	rows, err = db.Query(blocksQuery, ticketID)
	if err != nil {
		log.Fatalf("Error querying blockers: %v", err)
	}

	hasBlockers := false
	unresolvedCount := 0
	fmt.Println("This ticket blocks:")
	for rows.Next() {
		var blockedID, blockedTitle, blockedStatus string
		if err := rows.Scan(&blockedID, &blockedTitle, &blockedStatus); err != nil {
			log.Fatalf("Error scanning blocker: %v", err)
		}
		hasBlockers = true

		// Count unresolved
		if !workflow.IsCompleted(blockedStatus) {
			unresolvedCount++
		}

		fmt.Printf("  • %s: %s (%s)\n", blockedID, truncate(blockedTitle, 40), blockedStatus)
	}
	rows.Close()

	if !hasBlockers {
		fmt.Println("  (none)")
	}
	fmt.Println()

	// Show summary
	isCompleted := workflow.IsCompleted(status)
	if isCompleted {
		fmt.Printf("Status: %s ✓\n", status)
	} else {
		fmt.Printf("Status: %s\n", status)
	}

	if hasBlockers {
		if unresolvedCount > 0 {
			fmt.Printf("Blocking %d ticket(s) (%d unresolved)\n", unresolvedCount, unresolvedCount)
		} else {
			fmt.Printf("Blocking %d ticket(s) (all resolved)\n", 0)
		}
	}
}

func init() {
	rootCmd.AddCommand(blockedCmd)
}
