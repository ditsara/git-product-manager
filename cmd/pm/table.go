package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Status color palette — applied to any cell whose column is the status column.
var statusColors = map[string]lipgloss.Color{
	"backlog":     lipgloss.Color("245"), // gray
	"todo":        lipgloss.Color("12"),  // bright blue
	"in-progress": lipgloss.Color("11"),  // yellow
	"done":        lipgloss.Color("10"),  // green
	"canceled":    lipgloss.Color("238"), // dark gray
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true)
	baseStyle   = lipgloss.NewStyle()
)

// TableColumn defines a column's header and minimum display width.
type TableColumn struct {
	Header string
	Width  int
}

// renderTable renders a borderless, optionally color-coded table.
//
// statusColIndex is the 0-based column index whose values are color-coded by
// status. Pass -1 to disable status coloring.
func renderTable(cols []TableColumn, rows [][]string, statusColIndex int) string {
	// Pad all cells and headers to their minimum column widths up front.
	// lipgloss/table v1 doesn't expose per-column width directly, so we
	// enforce widths by padding values.
	paddedHeaders := make([]string, len(cols))
	for i, c := range cols {
		paddedHeaders[i] = padRight(c.Header, c.Width)
	}

	paddedRows := make([][]string, len(rows))
	for i, r := range rows {
		paddedRows[i] = make([]string, len(r))
		for j, cell := range r {
			minW := 0
			if j < len(cols) {
				minW = cols[j].Width
			}
			paddedRows[i][j] = padRight(cell, minW)
		}
	}

	styleFunc := func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return headerStyle
		}
		if statusColIndex >= 0 && col == statusColIndex && row < len(rows) {
			status := rows[row][statusColIndex]
			if color, ok := statusColors[status]; ok {
				return baseStyle.Foreground(color)
			}
		}
		return baseStyle
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		BorderHeader(false).
		StyleFunc(styleFunc).
		Headers(paddedHeaders...)

	for _, r := range paddedRows {
		t.Row(r...)
	}

	return t.Render()
}

// padRight pads s with spaces on the right to at least minWidth rune-length.
func padRight(s string, minWidth int) string {
	runes := []rune(s)
	if len(runes) >= minWidth {
		return s
	}
	for len(runes) < minWidth {
		runes = append(runes, ' ')
	}
	return string(runes)
}

