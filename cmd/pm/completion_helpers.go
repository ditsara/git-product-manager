package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/config"
	"github.com/ditsara/git-product-manager/internal/guide"
	"github.com/spf13/cobra"
)

// completeTicketIDs provides completion for ticket IDs
func completeTicketIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ticketsPath := filepath.Join(".pm", "tickets")
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		// Gracefully handle missing .pm directory
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var tickets []string
	prefix := strings.ToUpper(toComplete) // Case-insensitive matching

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		ticketID := strings.TrimSuffix(file.Name(), ".md")

		// Filter by prefix if provided
		if toComplete == "" || strings.HasPrefix(strings.ToUpper(ticketID), prefix) {
			tickets = append(tickets, ticketID)
		}
	}

	return tickets, cobra.ShellCompDirectiveNoFileComp
}

// completeStates provides completion for workflow states
func completeStates(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	workflowPath := filepath.Join(".pm", "config", "workflow.yaml")
	workflow, err := config.LoadWorkflow(workflowPath)
	if err != nil {
		// Gracefully handle missing workflow.yaml
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var states []string
	for _, state := range workflow.States {
		if strings.HasPrefix(state, toComplete) {
			states = append(states, state)
		}
	}

	return states, cobra.ShellCompDirectiveNoFileComp
}

// completeMilestoneIDs provides completion for milestone IDs
func completeMilestoneIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	milestonesPath := filepath.Join(".pm", "milestones")
	files, err := os.ReadDir(milestonesPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var ids []string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(file.Name(), ".md")
		if toComplete == "" || strings.HasPrefix(id, toComplete) {
			ids = append(ids, id)
		}
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

// completeTicketTypes provides completion for ticket types
func completeTicketTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	types := []string{"story", "task", "bug", "epic"}
	var matches []string

	for _, t := range types {
		if strings.HasPrefix(t, toComplete) {
			matches = append(matches, t)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

// completeMembers provides completion for assignee usernames from the members
// list in project.yaml. Returns no suggestions if the list is absent or empty.
func completeMembers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	pmPath := ".pm"
	project, err := config.LoadProject(pmPath)
	if err != nil || len(project.Members) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var matches []string
	for _, m := range project.Members {
		if toComplete == "" || strings.HasPrefix(m, toComplete) {
			matches = append(matches, m)
		}
	}
	return matches, cobra.ShellCompDirectiveNoFileComp
}

// completeGuideSections provides completion for pm guide section names.
func completeGuideSections(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var matches []string
	for _, s := range guide.SectionNames() {
		if strings.HasPrefix(s, toComplete) {
			matches = append(matches, s)
		}
	}
	return matches, cobra.ShellCompDirectiveNoFileComp
}
