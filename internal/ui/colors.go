package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/ditsara/git-product-manager/internal/config"
)

// GroupColors maps workflow state group names to terminal colors.
// States belonging to a group not listed here render with the base style.
var GroupColors = map[string]lipgloss.Color{
	"active":    lipgloss.Color("11"),  // yellow — e.g. todo, in-progress
	"completed": lipgloss.Color("238"), // dark gray — e.g. done, canceled
}

// TypeColors maps ticket type names to terminal colors.
var TypeColors = map[string]lipgloss.Color{
	"epic":  lipgloss.Color("21"),  // bold blue
	"story": lipgloss.Color("33"),  // blue
	"task":  lipgloss.Color("34"),  // green
	"bug":   lipgloss.Color("196"), // red
}

// StatusStyleFunc returns a closure that maps a status string to a lipgloss
// style, derived from the state groups defined in the workflow. States not
// belonging to a group in GroupColors render with the base style.
func StatusStyleFunc(w *config.Workflow) func(string) lipgloss.Style {
	// Build a reverse index: state → group name.
	stateGroup := make(map[string]string)
	for group, states := range w.StateGroups {
		for _, s := range states {
			stateGroup[s] = group
		}
	}

	base := lipgloss.NewStyle()
	return func(status string) lipgloss.Style {
		group, ok := stateGroup[status]
		if !ok {
			return base
		}
		if c, ok := GroupColors[group]; ok {
			return base.Foreground(c)
		}
		return base
	}
}

// TypeStyleFunc returns a closure that maps a ticket type string to a lipgloss
// style. Epics are rendered bold. Unknown types render with the base style.
func TypeStyleFunc() func(string) lipgloss.Style {
	base := lipgloss.NewStyle()
	return func(ticketType string) lipgloss.Style {
		c, ok := TypeColors[ticketType]
		if !ok {
			return base
		}
		s := base.Foreground(c)
		if ticketType == "epic" {
			s = s.Bold(true)
		}
		return s
	}
}

// TypeStyle returns a lipgloss style for the given ticket type directly.
// Kept for use outside of renderTable (e.g. pm tree's formatNode).
func TypeStyle(ticketType string, base lipgloss.Style) lipgloss.Style {
	if c, ok := TypeColors[ticketType]; ok {
		s := base.Foreground(c)
		if ticketType == "epic" {
			s = s.Bold(true)
		}
		return s
	}
	return base
}
