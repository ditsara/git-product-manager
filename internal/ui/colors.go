package ui

import "github.com/charmbracelet/lipgloss"

// StatusColors maps workflow state names to terminal colors.
var StatusColors = map[string]lipgloss.Color{
	"backlog":     lipgloss.Color("245"), // gray
	"todo":        lipgloss.Color("12"),  // bright blue
	"in-progress": lipgloss.Color("11"),  // yellow
	"done":        lipgloss.Color("10"),  // green
	"canceled":    lipgloss.Color("238"), // dark gray
}

// TypeColors maps ticket type names to terminal colors.
var TypeColors = map[string]lipgloss.Color{
	"epic":  lipgloss.Color("21"),  // bold blue
	"story": lipgloss.Color("33"),  // blue
	"task":  lipgloss.Color("34"),  // green
	"bug":   lipgloss.Color("196"), // red
}

// StatusStyle returns a lipgloss style for the given status, or the base style
// if the status is not recognised.
func StatusStyle(status string, base lipgloss.Style) lipgloss.Style {
	if c, ok := StatusColors[status]; ok {
		return base.Foreground(c)
	}
	return base
}

// TypeStyle returns a lipgloss style for the given ticket type, or the base
// style if the type is not recognised. Epics are rendered bold.
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
