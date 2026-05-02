package milestone

import (
	"fmt"
	"math"
	"time"
)

// ProgressInfo holds computed progress metrics for a milestone.
type ProgressInfo struct {
	TotalTickets  int
	DoneTickets   int
	TotalPoints   int
	DonePoints    int
	DaysRemaining int // negative = overdue; 0 = due today; math.MaxInt if no due date
	IsOverdue     bool
	HasDueDate    bool
}

// TicketSummary is the minimal ticket data needed for progress calculation and display.
type TicketSummary struct {
	ID     string
	Title  string
	Type   string
	Status string
	Points int
}

// CalculateProgress computes progress metrics from a list of ticket summaries
// and an optional due date string ("YYYY-MM-DD" or "").
func CalculateProgress(tickets []TicketSummary, dueDate string, doneStates []string) ProgressInfo {
	doneSet := make(map[string]bool, len(doneStates))
	for _, s := range doneStates {
		doneSet[s] = true
	}

	var info ProgressInfo
	info.TotalTickets = len(tickets)

	for _, t := range tickets {
		info.TotalPoints += t.Points
		if doneSet[t.Status] {
			info.DoneTickets++
			info.DonePoints += t.Points
		}
	}

	if dueDate == "" {
		info.HasDueDate = false
		info.DaysRemaining = math.MaxInt
		info.IsOverdue = false
	} else {
		due, err := time.Parse("2006-01-02", dueDate)
		if err != nil {
			info.HasDueDate = false
			info.DaysRemaining = 0
			info.IsOverdue = false
		} else {
			info.HasDueDate = true
			today := time.Now().UTC().Truncate(24 * time.Hour)
			info.DaysRemaining = int(due.Sub(today) / (24 * time.Hour))
			info.IsOverdue = info.DaysRemaining < 0
		}
	}

	return info
}

// ProgressBar renders an ASCII progress bar.
// Format: [████████░░░░░░░░] 50% (4/8)
// width is the number of bar characters (default 20 if <= 0).
func ProgressBar(done, total, width int) string {
	if width <= 0 {
		width = 20
	}

	var pct int
	var filled int
	if total > 0 {
		pct = done * 100 / total
		filled = done * width / total
	}

	bar := make([]rune, width)
	for i := 0; i < width; i++ {
		if i < filled {
			bar[i] = '█'
		} else {
			bar[i] = '░'
		}
	}

	return fmt.Sprintf("[%s] %d%% (%d/%d)", string(bar), pct, done, total)
}
