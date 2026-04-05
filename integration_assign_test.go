package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ditsara/git-product-manager/internal/config"
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

// --- GPM-29: assignee_domain ---

// TestAssignDomainAppended verifies domain is auto-appended when configured.
func TestAssignDomainAppended(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")
runPM(t, pmBinary, workspace, "new", "Test Ticket")

// Set domain in project.yaml
projectYAML := filepath.Join(workspace, ".pm", "config", "project.yaml")
if err := os.WriteFile(projectYAML, []byte("prefix: ASSIGN\nassignee_domain: \"example.com\"\n"), 0644); err != nil {
t.Fatalf("write project.yaml: %v", err)
}

output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "alice")
if err != nil {
t.Fatalf("pm assign failed: %v\nOutput: %s", err, output)
}

if !strings.Contains(output, "alice@example.com") {
t.Errorf("expected domain-appended assignee in output, got: %s", output)
}

ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
content, _ := os.ReadFile(ticketPath)
parsedTicket, _ := ticket.Parse(content)
if parsedTicket.Assignee != "alice@example.com" {
t.Errorf("expected assignee=alice@example.com, got %q", parsedTicket.Assignee)
}
}

// TestAssignDomainNotDoubleAppended verifies full emails are unchanged.
func TestAssignDomainNotDoubleAppended(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")
runPM(t, pmBinary, workspace, "new", "Test Ticket")

projectYAML := filepath.Join(workspace, ".pm", "config", "project.yaml")
if err := os.WriteFile(projectYAML, []byte("prefix: ASSIGN\nassignee_domain: \"example.com\"\n"), 0644); err != nil {
t.Fatalf("write project.yaml: %v", err)
}

output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "alice@other.org")
if err != nil {
t.Fatalf("pm assign failed: %v\nOutput: %s", err, output)
}

ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
content, _ := os.ReadFile(ticketPath)
parsedTicket, _ := ticket.Parse(content)
if parsedTicket.Assignee != "alice@other.org" {
t.Errorf("expected assignee unchanged at alice@other.org, got %q", parsedTicket.Assignee)
}
}

// TestAssignNoDomainUnchanged verifies no domain config means no appending.
func TestAssignNoDomainUnchanged(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")
runPM(t, pmBinary, workspace, "new", "Test Ticket")

output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "alice")
if err != nil {
t.Fatalf("pm assign failed: %v\nOutput: %s", err, output)
}

ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
content, _ := os.ReadFile(ticketPath)
parsedTicket, _ := ticket.Parse(content)
if parsedTicket.Assignee != "alice" {
t.Errorf("expected assignee=alice (no domain), got %q", parsedTicket.Assignee)
}
}

// --- GPM-79: members list ---

// TestProjectYAMLMembersField verifies the members field round-trips through config.
func TestProjectYAMLMembersField(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")

projectYAML := filepath.Join(workspace, ".pm", "config", "project.yaml")
if err := os.WriteFile(projectYAML, []byte("prefix: ASSIGN\nmembers:\n  - alice\n  - bob\n"), 0644); err != nil {
t.Fatalf("write project.yaml: %v", err)
}

// Load via config package
pmPath := filepath.Join(workspace, ".pm")
project, err := config.LoadProject(pmPath)
if err != nil {
t.Fatalf("LoadProject: %v", err)
}
if len(project.Members) != 2 || project.Members[0] != "alice" || project.Members[1] != "bob" {
t.Errorf("unexpected members: %v", project.Members)
}
}

// TestInitProjectYAMLHasCommentedMembers verifies pm init writes the members comment.
func TestInitProjectYAMLHasCommentedMembers(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")

projectYAML := filepath.Join(workspace, ".pm", "config", "project.yaml")
content, err := os.ReadFile(projectYAML)
if err != nil {
t.Fatalf("read project.yaml: %v", err)
}

if !strings.Contains(string(content), "members") {
t.Errorf("expected members comment in project.yaml, got:\n%s", content)
}
if !strings.Contains(string(content), "assignee_domain") {
t.Errorf("expected assignee_domain comment in project.yaml, got:\n%s", content)
}
}

// --- GPM-80: warn on unknown assignee ---

// TestAssignWarnsForUnknownMember verifies warning on stderr when not in members list.
func TestAssignWarnsForUnknownMember(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")
runPM(t, pmBinary, workspace, "new", "Test Ticket")

projectYAML := filepath.Join(workspace, ".pm", "config", "project.yaml")
if err := os.WriteFile(projectYAML, []byte("prefix: ASSIGN\nmembers:\n  - alice\n  - bob\n"), 0644); err != nil {
t.Fatalf("write project.yaml: %v", err)
}

// runPM uses CombinedOutput so stderr is in output
output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "charlie")
if err != nil {
t.Fatalf("pm assign failed: %v\nOutput: %s", err, output)
}

if !strings.Contains(output, "Warning") || !strings.Contains(output, "charlie") {
t.Errorf("expected warning about charlie not in member list, got: %s", output)
}

// Assignment should still succeed
ticketPath := filepath.Join(workspace, ".pm", "tickets", "ASSIGN-1.md")
content, _ := os.ReadFile(ticketPath)
parsedTicket, _ := ticket.Parse(content)
if parsedTicket.Assignee != "charlie" {
t.Errorf("expected assignee=charlie despite warning, got %q", parsedTicket.Assignee)
}
}

// TestAssignNoWarnForListedMember verifies no warning when assignee is in members list.
func TestAssignNoWarnForListedMember(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")
runPM(t, pmBinary, workspace, "new", "Test Ticket")

projectYAML := filepath.Join(workspace, ".pm", "config", "project.yaml")
if err := os.WriteFile(projectYAML, []byte("prefix: ASSIGN\nmembers:\n  - alice\n  - bob\n"), 0644); err != nil {
t.Fatalf("write project.yaml: %v", err)
}

output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "alice")
if err != nil {
t.Fatalf("pm assign failed: %v\nOutput: %s", err, output)
}

if strings.Contains(output, "Warning") {
t.Errorf("expected no warning for known member, got: %s", output)
}
}

// TestAssignNoWarnWithEmptyMembers verifies no warning when members list is empty.
func TestAssignNoWarnWithEmptyMembers(t *testing.T) {
workspace := t.TempDir()
pmBinary := buildPMBinary(t)

initWorkspace(t, pmBinary, workspace, "ASSIGN")
runPM(t, pmBinary, workspace, "new", "Test Ticket")

output, err := runPM(t, pmBinary, workspace, "assign", "ASSIGN-1", "anyone")
if err != nil {
t.Fatalf("pm assign failed: %v\nOutput: %s", err, output)
}

if strings.Contains(output, "Warning") {
t.Errorf("expected no warning when members list absent, got: %s", output)
}
}
