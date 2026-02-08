package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/config"
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
