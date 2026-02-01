package ticket

import (
	"strings"
	"testing"
	"time"
)

func TestTicketValidate(t *testing.T) {
	tests := []struct {
		name    string
		ticket  Ticket
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ticket",
			ticket: Ticket{
				ID:        "TEST-1",
				Title:     "Test ticket",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			ticket: Ticket{
				Title:     "Test ticket",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name: "invalid id format - lowercase prefix",
			ticket: Ticket{
				ID:        "test-1",
				Title:     "Test ticket",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: true,
			errMsg:  "invalid id format",
		},
		{
			name: "invalid id format - no number",
			ticket: Ticket{
				ID:        "TEST-ABC",
				Title:     "Test ticket",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: true,
			errMsg:  "invalid id format",
		},
		{
			name: "valid id with multiple parts",
			ticket: Ticket{
				ID:        "PROJ2024-123",
				Title:     "Test ticket",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			ticket: Ticket{
				ID:        "TEST-1",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "missing type",
			ticket: Ticket{
				ID:        "TEST-1",
				Title:     "Test ticket",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing status",
			ticket: Ticket{
				ID:        "TEST-1",
				Title:     "Test ticket",
				Type:      "story",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: true,
			errMsg:  "status is required",
		},
		{
			name: "invalid created_at format",
			ticket: Ticket{
				ID:        "TEST-1",
				Title:     "Test ticket",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31",
				UpdatedAt: "2026-01-31T09:00:00Z",
			},
			wantErr: true,
			errMsg:  "invalid created_at format",
		},
		{
			name: "invalid updated_at format",
			ticket: Ticket{
				ID:        "TEST-1",
				Title:     "Test ticket",
				Type:      "story",
				Status:    "backlog",
				CreatedAt: "2026-01-31T09:00:00Z",
				UpdatedAt: "not a date",
			},
			wantErr: true,
			errMsg:  "invalid updated_at format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ticket.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error containing '%s', got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing '%s'", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTicketParse(t *testing.T) {
	validYAML := `id: TEST-123
title: "Test Ticket"
type: story
status: backlog
priority: high
points: 5
parent: ""
depends_on: []
blocks: []
related: []
labels: [auth, api]
assignee: alice
created_at: 2026-01-31T09:00:00Z
updated_at: 2026-01-31T09:30:00Z`

	ticket, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("Parse() unexpected error = %v", err)
	}

	if ticket.ID != "TEST-123" {
		t.Errorf("Parse() ID = %v, want TEST-123", ticket.ID)
	}
	if ticket.Title != "Test Ticket" {
		t.Errorf("Parse() Title = %v, want 'Test Ticket'", ticket.Title)
	}
	if ticket.Type != "story" {
		t.Errorf("Parse() Type = %v, want 'story'", ticket.Type)
	}
	if ticket.Status != "backlog" {
		t.Errorf("Parse() Status = %v, want 'backlog'", ticket.Status)
	}
	if ticket.Priority != "high" {
		t.Errorf("Parse() Priority = %v, want 'high'", ticket.Priority)
	}
	if ticket.Points != 5 {
		t.Errorf("Parse() Points = %v, want 5", ticket.Points)
	}
	if ticket.Assignee != "alice" {
		t.Errorf("Parse() Assignee = %v, want 'alice'", ticket.Assignee)
	}
	if len(ticket.Labels) != 2 || ticket.Labels[0] != "auth" || ticket.Labels[1] != "api" {
		t.Errorf("Parse() Labels = %v, want [auth api]", ticket.Labels)
	}

	// Verify timestamps are valid
	if _, err := time.Parse(time.RFC3339, ticket.CreatedAt); err != nil {
		t.Errorf("Parse() CreatedAt invalid: %v", err)
	}
	if _, err := time.Parse(time.RFC3339, ticket.UpdatedAt); err != nil {
		t.Errorf("Parse() UpdatedAt invalid: %v", err)
	}
}

func TestTicketParseInvalidYAML(t *testing.T) {
	invalidYAML := `this is not valid YAML: [[[`

	_, err := Parse([]byte(invalidYAML))
	if err == nil {
		t.Error("Parse() expected error for invalid YAML, got nil")
	}
}

func TestTicketIDPattern(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid simple", "TEST-1", true},
		{"valid with number in prefix", "PROJ2024-123", true},
		{"valid long number", "ABC-999999", true},
		{"invalid lowercase", "test-1", false},
		{"invalid no dash", "TEST1", false},
		{"invalid no number", "TEST-", false},
		{"invalid letters after dash", "TEST-ABC", false},
		{"invalid starts with number", "1TEST-123", false},
		{"invalid empty", "", false},
		{"valid alphanumeric prefix", "TEST123-456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ticketIDPattern.MatchString(tt.id)
			if result != tt.valid {
				t.Errorf("ticketIDPattern.MatchString(%q) = %v, want %v", tt.id, result, tt.valid)
			}
		})
	}
}
