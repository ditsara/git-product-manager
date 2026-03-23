package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/config"
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
  pm milestone list --state closed
  pm milestone list --overdue
  pm milestone list --with-progress`,
	Run: func(cmd *cobra.Command, args []string) {
		stateFilter, _ := cmd.Flags().GetString("state")
		overdueOnly, _ := cmd.Flags().GetBool("overdue")
		withProgress, _ := cmd.Flags().GetBool("with-progress")
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

		// Load workflow for done states (used by overdue/progress flags)
		var doneStates []string
		if overdueOnly || withProgress {
			doneStates = loadDoneStates(pmPath)
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

		// Filter to overdue active milestones
		if overdueOnly {
			today := time.Now().UTC().Truncate(24 * time.Hour)
			var filtered []*milestone.Milestone
			for _, m := range milestones {
				if m.State != "active" || m.DueDate == "" {
					continue
				}
				due, err := time.Parse("2006-01-02", m.DueDate)
				if err != nil {
					continue
				}
				if due.Before(today) {
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

		ticketsDir := filepath.Join(pmPath, "tickets")

		if overdueOnly {
			today := time.Now().UTC().Truncate(24 * time.Hour)
			fmt.Printf("%-20s %-40s %-13s %-12s\n", "ID", "TITLE", "DUE DATE", "DAYS OVERDUE")
			fmt.Println(strings.Repeat("-", 85))
			for _, m := range milestones {
				dueDateStr := m.DueDate
				daysOverdue := ""
				if due, err := time.Parse("2006-01-02", m.DueDate); err == nil {
					dueDateStr = due.Format("Jan 02, 2006")
					days := int(today.Sub(due) / (24 * time.Hour))
					daysOverdue = fmt.Sprintf("%d", days)
				}
				fmt.Printf("%-20s %-40s %-13s %-12s\n",
					truncate(m.ID, 20),
					truncate(m.Title, 40),
					dueDateStr,
					daysOverdue,
				)
			}
			return
		}

		if withProgress {
			fmt.Printf("%-20s %-30s %-13s %-10s %-14s\n", "ID", "TITLE", "DUE DATE", "STATE", "PROGRESS")
			fmt.Println(strings.Repeat("-", 87))
			for _, m := range milestones {
				dueDateStr := "-"
				if m.DueDate != "" {
					if t, err := time.Parse("2006-01-02", m.DueDate); err == nil {
						dueDateStr = t.Format("Jan 02, 2006")
					} else {
						dueDateStr = m.DueDate
					}
				}
				summaries := collectTicketSummaries(ticketsDir, m.ID)
				info := milestone.CalculateProgress(summaries, m.DueDate, doneStates)
				progressStr := fmt.Sprintf("%d%% (%d/%d)", pct(info.DoneTickets, info.TotalTickets), info.DoneTickets, info.TotalTickets)
				fmt.Printf("%-20s %-30s %-13s %-10s %-14s\n",
					truncate(m.ID, 20),
					truncate(m.Title, 30),
					dueDateStr,
					m.State,
					progressStr,
				)
			}
			return
		}

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

		// Progress section
		ticketsDir := filepath.Join(pmPath, "tickets")
		doneStates := loadDoneStates(pmPath)
		summaries := collectTicketSummaries(ticketsDir, m.ID)
		info := milestone.CalculateProgress(summaries, m.DueDate, doneStates)

		fmt.Println()
		if info.TotalTickets == 0 {
			fmt.Printf("%-12s 0 (none assigned)\n", "Tickets:")
		} else {
			bar := milestone.ProgressBar(info.DoneTickets, info.TotalTickets, 20)
			fmt.Printf("%-12s %s tickets\n", "Progress:", bar)

			if info.TotalPoints > 0 {
				pointsBar := milestone.ProgressBar(info.DonePoints, info.TotalPoints, 20)
				fmt.Printf("%-12s %s points\n", "By Points:", pointsBar)
			}

			if info.HasDueDate {
				if due, err := time.Parse("2006-01-02", m.DueDate); err == nil {
					if info.IsOverdue {
						days := -info.DaysRemaining
						fmt.Printf("%-12s ⚠ OVERDUE: Due %s (%d days ago)\n", "Due:", due.Format("Jan 02, 2006"), days)
					} else {
						fmt.Printf("%-12s %s (%d days)\n", "Due:", due.Format("Jan 02, 2006"), info.DaysRemaining)
					}
				}
			}
		}

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

// collectTicketSummaries returns TicketSummary for all tickets assigned to milestoneID.
func collectTicketSummaries(ticketsDir, milestoneID string) []milestone.TicketSummary {
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return nil
	}

	var summaries []milestone.TicketSummary
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
				summaries = append(summaries, milestone.TicketSummary{
					Status: t.Status,
					Points: t.Points,
				})
				break
			}
		}
	}
	return summaries
}

// loadDoneStates loads the "completed" state group from workflow.yaml,
// falling back to ["done", "canceled"] if the config is missing or malformed.
func loadDoneStates(pmPath string) []string {
	workflowPath := filepath.Join(pmPath, "config", "workflow.yaml")
	wf, err := config.LoadWorkflow(workflowPath)
	if err != nil {
		return []string{"done", "canceled"}
	}
	completed := wf.GetCompletedStates()
	if len(completed) == 0 {
		return []string{"done", "canceled"}
	}
	return completed
}

// pct computes integer percentage (0 if total == 0).
func pct(done, total int) int {
	if total == 0 {
		return 0
	}
	return done * 100 / total
}

// milestoneCloseCmd closes a milestone.
var milestoneCloseCmd = &cobra.Command{
	Use:               "close <milestone-id>",
	Short:             "Close a milestone",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeMilestoneIDs,
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")
		pmPath := ".pm"
		milestonesDir := filepath.Join(pmPath, "milestones")

		m, err := findMilestone(milestonesDir, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if m.State != "active" {
			fmt.Printf("Milestone '%s' is already closed.\n", m.ID)
			return
		}

		doneStates := loadDoneStates(pmPath)
		ticketsDir := filepath.Join(pmPath, "tickets")
		summaries := collectTicketSummaries(ticketsDir, m.ID)
		info := milestone.CalculateProgress(summaries, m.DueDate, doneStates)

		incomplete := info.TotalTickets - info.DoneTickets
		if incomplete > 0 && !force {
			fmt.Fprintf(os.Stderr, "Error: %d ticket(s) are not done. Use --force to close anyway.\n", incomplete)

			// Collect incomplete ticket info for display
			type incompleteTicket struct {
				id    string
				title string
			}
			var incompletes []incompleteTicket

			doneSet := make(map[string]bool, len(doneStates))
			for _, s := range doneStates {
				doneSet[s] = true
			}

			entries, _ := os.ReadDir(ticketsDir)
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
				if doneSet[t.Status] {
					continue
				}
				for _, mid := range t.Milestones {
					if mid == m.ID {
						incompletes = append(incompletes, incompleteTicket{id: t.ID, title: t.Title})
						break
					}
				}
			}

			shown := incompletes
			extra := 0
			if len(shown) > 5 {
				extra = len(shown) - 5
				shown = shown[:5]
			}
			for _, inc := range shown {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", inc.id, inc.title)
			}
			if extra > 0 {
				fmt.Fprintf(os.Stderr, "  ... and %d more\n", extra)
			}
			os.Exit(1)
		}

		// Close the milestone
		m.State = "closed"
		m.ClosedAt = time.Now().UTC().Format(time.RFC3339)

		destPath := filepath.Join(milestonesDir, m.ID+".md")
		if err := milestone.Write(m, destPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		commitMsg := fmt.Sprintf("chore(pm): Close milestone %s", m.ID)
		gitAdd := exec.Command("git", "add", destPath)
		gitAdd.Dir = "."
		_ = gitAdd.Run()

		gitCommit := exec.Command("git", "commit", "-m", commitMsg)
		gitCommit.Dir = "."
		_ = gitCommit.Run()

		fmt.Printf("✓ Closed milestone: %s\n", m.ID)
		fmt.Printf("  %d/%d tickets done\n", info.DoneTickets, info.TotalTickets)
	},
}

func init() {
	// Create flags for milestone create
	milestoneCreateCmd.Flags().String("due", "", "Due date in YYYY-MM-DD format")
	milestoneCreateCmd.Flags().String("description", "", "Short description of the milestone")
	milestoneCreateCmd.Flags().String("id", "", "Custom milestone ID (slug format, e.g. v1-release)")

	// State filter for list
	milestoneListCmd.Flags().String("state", "", "Filter by state (active, closed)")
	milestoneListCmd.Flags().Bool("overdue", false, "Show only overdue active milestones")
	milestoneListCmd.Flags().Bool("with-progress", false, "Show completion percentage")

	// Flags for close
	milestoneCloseCmd.Flags().Bool("force", false, "Close even if tickets are not done")

	milestoneCmd.AddCommand(milestoneCreateCmd)
	milestoneCmd.AddCommand(milestoneListCmd)
	milestoneCmd.AddCommand(milestoneShowCmd)
	milestoneCmd.AddCommand(milestoneCloseCmd)

	rootCmd.AddCommand(milestoneCmd)
}
