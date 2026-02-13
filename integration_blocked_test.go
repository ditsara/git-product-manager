package main

import (
	"strings"
	"testing"
	"time"
)

func TestBlockedCommand(t *testing.T) {
	// Build the binary
	pmBinary := buildPMBinary(t)

	// Create test workspace
	workspace := t.TempDir()

	// Initialize pm
	initWorkspace(t, pmBinary, workspace, "TEST")

	// Create test tickets with dependencies
	// Ticket 1: No dependencies
	output, err := runPM(t, pmBinary, workspace, "new", "Base ticket")
	if err != nil {
		t.Fatalf("Failed to create ticket 1: %v\n%s", err, output)
	}
	// Extract ticket ID from "✓ Created new ticket: TEST-1"
	parts := strings.Fields(output)
	ticket1ID := parts[len(parts)-1]

	// Ticket 2: No dependencies  
	output, err = runPM(t, pmBinary, workspace, "new", "Another base ticket")
	if err != nil {
		t.Fatalf("Failed to create ticket 2: %v\n%s", err, output)
	}
	parts = strings.Fields(output)
	ticket2ID := parts[len(parts)-1]

	// Ticket 3: Depends on ticket 1
	output, err = runPM(t, pmBinary, workspace, "new", "Dependent ticket")
	if err != nil {
		t.Fatalf("Failed to create ticket 3: %v\n%s", err, output)
	}
	parts = strings.Fields(output)
	ticket3ID := parts[len(parts)-1]

	// Link ticket 3 to depend on ticket 1
	output, err = runPM(t, pmBinary, workspace, "link", ticket3ID, ticket1ID, "--type", "depends-on")
	if err != nil {
		t.Fatalf("Failed to link tickets: %v\n%s", err, output)
	}

	// Ticket 4: Depends on ticket 2 (which we'll mark as done)
	output, err = runPM(t, pmBinary, workspace, "new", "Another dependent ticket")
	if err != nil {
		t.Fatalf("Failed to create ticket 4: %v\n%s", err, output)
	}
	parts = strings.Fields(output)
	ticket4ID := parts[len(parts)-1]

	// Link ticket 4 to depend on ticket 2
	output, err = runPM(t, pmBinary, workspace, "link", ticket4ID, ticket2ID, "--type", "depends-on")
	if err != nil {
		t.Fatalf("Failed to link tickets: %v\n%s", err, output)
	}

	// Move ticket 2 to done
	output, err = runPM(t, pmBinary, workspace, "move", ticket2ID, "done")
	if err != nil {
		t.Fatalf("Failed to move ticket: %v\n%s", err, output)
	}

	// Sleep to ensure file modification time is in a different second
	// (filesystem mtime has 1-second precision on most systems)
	// Increased to 2s to ensure we cross a second boundary even if the
	// test starts near the end of a second
	time.Sleep(2100 * time.Millisecond)

	// Test 1: Global blocked view should show ticket 3 (blocked by backlog ticket 1)
	// but NOT ticket 4 (blocked by done ticket 2)
	output, err = runPM(t, pmBinary, workspace, "blocked")
	if err != nil {
		t.Fatalf("Failed to run blocked: %v\n%s", err, output)
	}

	if !strings.Contains(output, ticket3ID) {
		t.Errorf("Expected to find ticket %s in blocked output, but didn't:\n%s", ticket3ID, output)
	}

	if strings.Contains(output, ticket4ID) {
		t.Errorf("Did not expect to find ticket %s in blocked output (dependency is done), but found it:\n%s", ticket4ID, output)
	}

	if !strings.Contains(output, "1 ticket(s) blocked by 1 dependenc(ies)") {
		t.Errorf("Expected summary line '1 ticket(s) blocked by 1 dependenc(ies)', got:\n%s", output)
	}

	// Test 2: Specific ticket view for ticket 3
	output, err = runPM(t, pmBinary, workspace, "blocked", ticket3ID)
	if err != nil {
		t.Fatalf("Failed to run blocked %s: %v\n%s", ticket3ID, err, output)
	}

	if !strings.Contains(output, ticket1ID) {
		t.Errorf("Expected to find dependency %s in output for %s:\n%s", ticket1ID, ticket3ID, output)
	}

	if !strings.Contains(output, "This ticket depends on:") {
		t.Errorf("Expected 'This ticket depends on:' in output:\n%s", output)
	}

	// Test 3: Specific ticket view for ticket 1 (which blocks ticket 3)
	output, err = runPM(t, pmBinary, workspace, "blocked", ticket1ID)
	if err != nil {
		t.Fatalf("Failed to run blocked %s: %v\n%s", ticket1ID, err, output)
	}

	if !strings.Contains(output, ticket3ID) {
		t.Errorf("Expected to find blocked ticket %s in output for %s:\n%s", ticket3ID, ticket1ID, output)
	}

	if !strings.Contains(output, "This ticket blocks:") {
		t.Errorf("Expected 'This ticket blocks:' in output:\n%s", output)
	}

	// Test 4: Ticket with resolved dependency should show checkmark
	output, err = runPM(t, pmBinary, workspace, "blocked", ticket4ID)
	if err != nil {
		t.Fatalf("Failed to run blocked %s: %v\n%s", ticket4ID, err, output)
	}

	if !strings.Contains(output, "✓") {
		t.Errorf("Expected to find checkmark for resolved dependency in output:\n%s", output)
	}

	// Test 5: No blocked tickets when all dependencies are resolved
	// Move ticket 1 to done
	output, err = runPM(t, pmBinary, workspace, "move", ticket1ID, "done")
	if err != nil {
		t.Fatalf("Failed to move ticket: %v\n%s", err, output)
	}

	// Sleep to ensure file modification time is updated
	// Increased to 2s to ensure we cross a second boundary
	time.Sleep(2100 * time.Millisecond)

	// Run blocked command - it should auto-sync and show no blocked tickets
	output, err = runPM(t, pmBinary, workspace, "blocked")
	if err != nil {
		t.Fatalf("Failed to run blocked: %v\n%s", err, output)
	}

	if !strings.Contains(output, "No blocked tickets found") {
		t.Errorf("Expected 'No blocked tickets found' when all dependencies resolved, got:\n%s", output)
	}
}

func TestBlockedNonExistentTicket(t *testing.T) {
	// Build the binary
	pmBinary := buildPMBinary(t)

	// Create test workspace
	workspace := t.TempDir()

	// Initialize pm
	initWorkspace(t, pmBinary, workspace, "TEST")

	// Try to show blocked view for non-existent ticket
	output, err := runPM(t, pmBinary, workspace, "blocked", "TEST-999")
	
	// Should fail with error
	if err == nil {
		t.Errorf("Expected error for non-existent ticket, but succeeded")
	}

	if !strings.Contains(output, "Ticket not found") {
		t.Errorf("Expected 'Ticket not found' error message, got: %s", output)
	}
}

