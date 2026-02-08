package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
)

// TestAssignTicket tests basic assignment
func TestAssignTicket(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "ASSIGN")
	runPM(t, pmBinary, workspace, "new", "Test Ticket")

	// Assign ticket
	output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "alice")
	if err != nil {
		t.Fatalf("pm assign failed: %v", err)
	}

	if !strings.Contains(output, "Assigned ASSIGN-1 to alice") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify assignee was updated
	ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	parsedTicket, err := ticket.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if parsedTicket.Assignee != "alice" {
		t.Errorf("Expected assignee=alice, got %q", parsedTicket.Assignee)
	}
}

// TestAssignTicketIdempotent tests that assigning same user shows message and doesn't update
func TestAssignTicketIdempotent(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "ASSIGN")
	runPM(t, pmBinary, workspace, "new", "Test Ticket")

	// Assign ticket first time
	runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "bob")

	ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	firstTicket, err := ticket.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	firstUpdated := firstTicket.UpdatedAt

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Assign same user again
	output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "bob")
	if err != nil {
		t.Fatalf("pm assign (idempotent) failed: %v", err)
	}

	if !strings.Contains(output, "Already assigned to bob") {
		t.Errorf("Expected idempotent message, got: %s", output)
	}

	// Verify nothing was updated
	content, err = os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	secondTicket, err := ticket.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if secondTicket.UpdatedAt != firstUpdated {
		t.Errorf("updated_at should not change on idempotent assignment")
	}
}

// TestAssignTicketWithEmail tests assigning ticket with email address
func TestAssignTicketWithEmail(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "ASSIGN")
	runPM(t, pmBinary, workspace, "new", "Test Ticket")

	// Assign with email
	output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "charlie@company.com")
	if err != nil {
		t.Fatalf("pm assign with email failed: %v", err)
	}

	if !strings.Contains(output, "Assigned ASSIGN-1 to charlie@company.com") {
		t.Errorf("Expected success message with email, got: %s", output)
	}

	// Verify email was stored
	ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	parsedTicket, err := ticket.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if parsedTicket.Assignee != "charlie@company.com" {
		t.Errorf("Expected assignee=charlie@company.com, got %q", parsedTicket.Assignee)
	}
}

// TestAssignTicketCaseInsensitive tests case-insensitive ticket ID matching
func TestAssignTicketCaseInsensitive(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "ASSIGN")
	runPM(t, pmBinary, workspace, "new", "Test Ticket")

	// Assign using lowercase ticket ID
	output, err := runPM(t, pmBinary, workspace, "assign", "assign-1", "dave")
	if err != nil {
		t.Fatalf("pm assign with lowercase ID failed: %v", err)
	}

	// Should show success for either case format
	if !strings.Contains(output, "Assigned") || !strings.Contains(output, "dave") {
		t.Errorf("Expected success with case-insensitive matching, got: %s", output)
	}

	// Verify correct ticket was updated
	ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	parsedTicket, err := ticket.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if parsedTicket.Assignee != "dave" {
		t.Errorf("Expected assignee=dave, got %q", parsedTicket.Assignee)
	}
}

// TestAssignTicketUpdateTimestamp tests that updated_at changes on assignment
func TestAssignTicketUpdateTimestamp(t *testing.T) {
	workspace := t.TempDir()
	pmBinary := buildPMBinary(t)

	initWorkspace(t, pmBinary, workspace, "ASSIGN")
	runPM(t, pmBinary, workspace, "new", "Test Ticket")

	ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	originalTicket, err := ticket.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	originalUpdated := originalTicket.UpdatedAt

	// Wait to ensure timestamp is different (timestamps have 1s resolution)
	time.Sleep(1100 * time.Millisecond)

	// Assign ticket
	runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "eve")

	// Verify updated_at changed
	content, err = os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	updatedTicket, err := ticket.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if updatedTicket.UpdatedAt == originalUpdated {
		t.Errorf("updated_at should change after assignment")
	}
}
