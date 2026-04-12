package main

import (
	"os"
	"strings"

	cterm "github.com/charmbracelet/x/term"
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

// termWidth returns the current terminal width, falling back to 100.
func termWidth() int {
	w, _, err := cterm.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		return 100
	}
	return w
}

// renderTable renders a borderless, optionally color-coded table.
//
// statusColIndex is the 0-based column index whose values are color-coded by
// status. Pass -1 to disable status coloring.
//
// expandCol is the 0-based column index that should absorb any spare terminal
// width (typically the TITLE column). Pass -1 to use fixed widths only.
func renderTable(cols []TableColumn, rows [][]string, statusColIndex, expandCol int) string {
	// Resolve effective column widths, expanding one column to fill the terminal.
	effective := make([]TableColumn, len(cols))
	copy(effective, cols)

	if expandCol >= 0 && expandCol < len(cols) {
		tw := termWidth()
		// Account for the single space lipgloss HiddenBorder adds between columns.
		separators := len(cols) - 1
		fixed := separators
		for i, c := range cols {
			if i != expandCol {
				fixed += c.Width
			}
		}
		expanded := tw - fixed
		minExpanded := 80 - fixed
		if minExpanded < 1 {
			minExpanded = 1
		}
		if expanded < minExpanded {
			expanded = minExpanded
		}
		effective[expandCol].Width = expanded
	}

	// Pad all cells and headers to their effective column widths.
	// lipgloss/table v1 doesn't expose per-column width directly, so we
	// enforce widths by padding (and truncating) values.
	paddedHeaders := make([]string, len(effective))
	for i, c := range effective {
		paddedHeaders[i] = padRight(c.Header, c.Width)
	}

	paddedRows := make([][]string, len(rows))
	for i, r := range rows {
		paddedRows[i] = make([]string, len(r))
		for j, cell := range r {
			w := 0
			if j < len(effective) {
				w = effective[j].Width
			}
			paddedRows[i][j] = padRight(truncate(cell, w-1), w)
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

// truncateID truncates a ticket ID to maxWidth runes, preserving the numeric
// suffix ("-NNNN" and optional " (+)") and using a single "…" marker so the ID
// remains identifiable. If the ID already fits, it is returned unchanged.
//
// Example (maxWidth=15):
//
//	"MYLONGPREFIX-1234"      → "MYLONG…-1234"
//	"MYLONGPREFIX-1234 (+)"  → "MYLO…-1234 (+)"
//	"GPM-42"                 → "GPM-42"
func truncateID(id string, maxWidth int) string {
	runes := []rune(id)
	if len(runes) <= maxWidth {
		return id
	}

	// Separate the " (+)" children indicator if present.
	childSuffix := ""
	bare := id
	if len(id) >= 4 && id[len(id)-4:] == " (+)" {
		childSuffix = " (+)"
		bare = id[:len(id)-4]
	}

	// Split on the last "-" to isolate the numeric part.
	lastDash := strings.LastIndex(bare, "-")
	if lastDash < 0 {
		// No dash — fall back to plain rune truncation with ellipsis.
		if maxWidth < 1 {
			return "…"
		}
		return string([]rune(id)[:maxWidth-1]) + "…"
	}

	numericSuffix := bare[lastDash:] + childSuffix // e.g. "-1234" + " (+)"
	// Available runes for the prefix: maxWidth minus suffix minus 1 for "…"
	available := maxWidth - len([]rune(numericSuffix)) - 1
	if available < 1 {
		available = 1
	}
	prefix := []rune(bare[:lastDash])
	if len(prefix) > available {
		prefix = prefix[:available]
	}
	return string(prefix) + "…" + numericSuffix
}

