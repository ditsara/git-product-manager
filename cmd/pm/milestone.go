package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/milestone"
	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
)

// milestoneCmd is the top-level milestone command group.
var milestoneCmd = &cobra.Command{
	Use:   "milestone",
	Short: "Manage milestones",
	Long:  "Create, list, and manage milestones for grouping tickets toward a goal.",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// milestoneCreateCmd creates a new milestone.
var milestoneCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new milestone",
	Long: `Create a new milestone file at .pm/milestones/{id}.md.

Examples:
  pm milestone create "Version 1.0"
  pm milestone create "Sprint 3" --due 2026-03-31
  pm milestone create "Sprint 3" --due 2026-03-31 --description "Q1 sprint"
  pm milestone create "Version 1.0" --id v1-release`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		dueDate, _ := cmd.Flags().GetString("due")
		description, _ := cmd.Flags().GetString("description")
		customID, _ := cmd.Flags().GetString("id")
		pmPath := ".pm"

		// Resolve the milestone ID
		var id string
		if customID != "" {
			if err := milestone.ValidateID(customID); err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid --id: %v\n", err)
				os.Exit(1)
			}
			id = customID
		} else {
			id = milestone.SlugFromTitle(title)
		}

		// Ensure milestones directory exists (lazy init for older projects)
		milestonesDir := filepath.Join(pmPath, "milestones")
		if err := os.MkdirAll(milestonesDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create milestones directory: %v\n", err)
			os.Exit(1)
		}

		// Check for ID collision
		destPath := filepath.Join(milestonesDir, id+".md")
		if _, err := os.Stat(destPath); err == nil {
			fmt.Fprintf(os.Stderr, "Error: Milestone ID '%s' already exists. Use --id to specify a unique ID.\n", id)
			os.Exit(1)
		}

		// Validate due date format if provided
		if dueDate != "" {
			if _, err := time.Parse("2006-01-02", dueDate); err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid --due date %q: expected format YYYY-MM-DD\n", dueDate)
				os.Exit(1)
			}
		}

		m := &milestone.Milestone{
			ID:          id,
			Title:       title,
			Description: description,
			DueDate:     dueDate,
			State:       "active",
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			ClosedAt:    "",
		}

		if err := milestone.Write(m, destPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Git stage the file
		gitAdd := exec.Command("git", "add", destPath)
		gitAdd.Dir = "."
		_ = gitAdd.Run() // non-fatal if not in a git repo

		fmt.Printf("✓ Created milestone: %s\n", id)
	},
}

// milestoneListCmd lists milestones.
var milestoneListCmd = &cobra.Command{
	Use:   "list",
	Short: "List milestones",
	Long: `List milestones with optional state filtering.

Examples:
  pm milestone list
  pm milestone list --state active
  pm milestone list --state closed`,
	Run: func(cmd *cobra.Command, args []string) {
		stateFilter, _ := cmd.Flags().GetString("state")
		pmPath := ".pm"
		milestonesDir := filepath.Join(pmPath, "milestones")

		milestones, err := milestone.ListMilestones(milestonesDir)
		if err != nil {
			if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
				fmt.Println("No milestones found.")
				return
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Filter by state if requested
		if stateFilter != "" {
			var filtered []*milestone.Milestone
			for _, m := range milestones {
				if m.State == stateFilter {
					filtered = append(filtered, m)
				}
			}
			milestones = filtered
		}

		if len(milestones) == 0 {
			fmt.Println("No milestones found.")
			return
		}

		// Sort: milestones with due dates ascending, then no-due-date ones at end
		sort.Slice(milestones, func(i, j int) bool {
			di, dj := milestones[i].DueDate, milestones[j].DueDate
			if di == "" && dj == "" {
				return milestones[i].ID < milestones[j].ID
			}
			if di == "" {
				return false
			}
			if dj == "" {
				return true
			}
			return di < dj
		})

		fmt.Printf("%-20s %-40s %-13s %-10s\n", "ID", "TITLE", "DUE DATE", "STATE")
		fmt.Println(strings.Repeat("-", 83))

		for _, m := range milestones {
			dueDateStr := "-"
			if m.DueDate != "" {
				if t, err := time.Parse("2006-01-02", m.DueDate); err == nil {
					dueDateStr = t.Format("Jan 02, 2006")
				} else {
					dueDateStr = m.DueDate
				}
			}
			fmt.Printf("%-20s %-40s %-13s %-10s\n",
				truncate(m.ID, 20),
				truncate(m.Title, 40),
				dueDateStr,
				m.State,
			)
		}
	},
}

// milestoneShowCmd shows a single milestone.
var milestoneShowCmd = &cobra.Command{
	Use:               "show <milestone-id>",
	Short:             "Show a milestone",
	Long:              `Show details of a single milestone.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeMilestoneIDs,
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		pmPath := ".pm"
		milestonesDir := filepath.Join(pmPath, "milestones")

		m, err := findMilestone(milestonesDir, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%-12s %s\n", "ID:", m.ID)
		fmt.Printf("%-12s %s\n", "Title:", m.Title)
		fmt.Printf("%-12s %s\n", "State:", m.State)

		if m.DueDate != "" {
			if t, err := time.Parse("2006-01-02", m.DueDate); err == nil {
				fmt.Printf("%-12s %s\n", "Due Date:", t.Format("Jan 02, 2006"))
			} else {
				fmt.Printf("%-12s %s\n", "Due Date:", m.DueDate)
			}
		}

		if m.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
				fmt.Printf("%-12s %s\n", "Created:", t.Format("Jan 02, 2006"))
			} else {
				fmt.Printf("%-12s %s\n", "Created:", m.CreatedAt)
			}
		}

		if m.ClosedAt != "" {
			if t, err := time.Parse(time.RFC3339, m.ClosedAt); err == nil {
				fmt.Printf("%-12s %s\n", "Closed:", t.Format("Jan 02, 2006"))
			} else {
				fmt.Printf("%-12s %s\n", "Closed:", m.ClosedAt)
			}
		}

		if m.Description != "" {
			fmt.Printf("%-12s %s\n", "Description:", m.Description)
		}

		// Count tickets associated with this milestone
		ticketCount := countTicketsForMilestone(filepath.Join(pmPath, "tickets"), m.ID)
		fmt.Printf("%-12s %d total\n", "Tickets:", ticketCount)

		if m.Body != "" {
			fmt.Println()
			fmt.Println(m.Body)
		}
	},
}

// findMilestone finds a milestone by exact match then prefix match.
func findMilestone(milestonesDir, id string) (*milestone.Milestone, error) {
	// Try exact match first
	exactPath := filepath.Join(milestonesDir, id+".md")
	if _, err := os.Stat(exactPath); err == nil {
		return milestone.ParseFile(exactPath)
	}

	// Scan for prefix match
	entries, err := os.ReadDir(milestonesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("milestone '%s' not found", id)
		}
		return nil, fmt.Errorf("failed to read milestones directory: %w", err)
	}

	var matches []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if strings.HasPrefix(e.Name(), id) {
			matches = append(matches, e.Name())
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("milestone '%s' not found", id)
	}
	if len(matches) > 1 {
		ids := make([]string, len(matches))
		for i, m := range matches {
			ids[i] = strings.TrimSuffix(m, ".md")
		}
		return nil, fmt.Errorf("ambiguous ID '%s'. Multiple matches: %s", id, strings.Join(ids, ", "))
	}

	return milestone.ParseFile(filepath.Join(milestonesDir, matches[0]))
}

// countTicketsForMilestone counts tickets whose Milestones field contains milestoneID.
func countTicketsForMilestone(ticketsDir, milestoneID string) int {
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(ticketsDir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		t, err := ticket.Parse(content)
		if err != nil {
			continue
		}
		for _, mid := range t.Milestones {
			if mid == milestoneID {
				count++
				break
			}
		}
	}
	return count
}

func init() {
	// Create flags for milestone create
	milestoneCreateCmd.Flags().String("due", "", "Due date in YYYY-MM-DD format")
	milestoneCreateCmd.Flags().String("description", "", "Short description of the milestone")
	milestoneCreateCmd.Flags().String("id", "", "Custom milestone ID (slug format, e.g. v1-release)")

	// State filter for list
	milestoneListCmd.Flags().String("state", "", "Filter by state (active, closed)")

	// TODO: pm milestone close is deferred to GPM-56

	milestoneCmd.AddCommand(milestoneCreateCmd)
	milestoneCmd.AddCommand(milestoneListCmd)
	milestoneCmd.AddCommand(milestoneShowCmd)

	rootCmd.AddCommand(milestoneCmd)
}
