package milestone

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestCalculateProgress_Basic(t *testing.T) {
	tickets := []TicketSummary{
		{Status: "done", Points: 1},
		{Status: "done", Points: 0},
		{Status: "done", Points: 1},
		{Status: "done", Points: 0},
		{Status: "todo", Points: 1},
		{Status: "in-progress", Points: 1},
		{Status: "backlog", Points: 1},
		{Status: "backlog", Points: 0},
		{Status: "todo", Points: 0},
		{Status: "todo", Points: 1},
	}
	info := CalculateProgress(tickets, "", []string{"done"})
	if info.TotalTickets != 10 {
		t.Errorf("TotalTickets: want 10, got %d", info.TotalTickets)
	}
	if info.DoneTickets != 4 {
		t.Errorf("DoneTickets: want 4, got %d", info.DoneTickets)
	}
	if info.TotalPoints != 6 {
		t.Errorf("TotalPoints: want 6, got %d", info.TotalPoints)
	}
	if info.DonePoints != 2 {
		t.Errorf("DonePoints: want 2, got %d", info.DonePoints)
	}
	if info.HasDueDate {
		t.Errorf("HasDueDate: want false, got true")
	}
	if info.IsOverdue {
		t.Errorf("IsOverdue: want false, got true")
	}
	if info.DaysRemaining != math.MaxInt {
		t.Errorf("DaysRemaining: want math.MaxInt, got %d", info.DaysRemaining)
	}
}

func TestCalculateProgress_FutureDueDate(t *testing.T) {
	future := time.Now().UTC().AddDate(0, 0, 10).Format("2006-01-02")
	info := CalculateProgress([]TicketSummary{{Status: "done", Points: 1}}, future, []string{"done"})
	if !info.HasDueDate {
		t.Errorf("HasDueDate: want true, got false")
	}
	if info.IsOverdue {
		t.Errorf("IsOverdue: want false, got true")
	}
	if info.DaysRemaining <= 0 {
		t.Errorf("DaysRemaining: want > 0, got %d", info.DaysRemaining)
	}
}

func TestCalculateProgress_PastDueDate(t *testing.T) {
	past := time.Now().UTC().AddDate(0, 0, -5).Format("2006-01-02")
	info := CalculateProgress([]TicketSummary{{Status: "todo", Points: 0}}, past, []string{"done"})
	if !info.HasDueDate {
		t.Errorf("HasDueDate: want true, got false")
	}
	if !info.IsOverdue {
		t.Errorf("IsOverdue: want true, got false")
	}
	if info.DaysRemaining >= 0 {
		t.Errorf("DaysRemaining: want < 0, got %d", info.DaysRemaining)
	}
}

func TestCalculateProgress_ZeroTickets(t *testing.T) {
	// Should not panic
	info := CalculateProgress([]TicketSummary{}, "", []string{"done"})
	if info.TotalTickets != 0 {
		t.Errorf("TotalTickets: want 0, got %d", info.TotalTickets)
	}
	if info.DoneTickets != 0 {
		t.Errorf("DoneTickets: want 0, got %d", info.DoneTickets)
	}
}

func TestProgressBar(t *testing.T) {
	cases := []struct {
		done, total, width int
		wantContains       []string
	}{
		{0, 0, 20, []string{"0%", "(0/0)", "[" + strings.Repeat("░", 20) + "]"}},
		{4, 8, 20, []string{"50%", "(4/8)"}},
		{8, 8, 20, []string{"100%", "(8/8)", "[" + strings.Repeat("█", 20) + "]"}},
	}

	for _, tc := range cases {
		got := ProgressBar(tc.done, tc.total, tc.width)
		for _, want := range tc.wantContains {
			if !strings.Contains(got, want) {
				t.Errorf("ProgressBar(%d,%d,%d) = %q, want to contain %q",
					tc.done, tc.total, tc.width, got, want)
			}
		}
	}
}
