package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper function to build the pm binary for testing
func buildPMBinary(t *testing.T) string {
	t.Helper()

	projectRoot := getProjectRoot(t)

	// Build in the project root's bin directory
	binDir := filepath.Join(projectRoot, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}
	binPath := filepath.Join(binDir, "pm")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/pm")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build pm binary: %v\nOutput: %s", err, output)
	}

	return binPath
}

// Helper function to get project root directory
func getProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current file's directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}

// Helper function to run pm command
func runPM(t *testing.T, pmBinary, workDir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(pmBinary, args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func initGitRepo(t *testing.T, workspace string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = workspace
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = workspace
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config user.name failed: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = workspace
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config user.email failed: %v\nOutput: %s", err, output)
	}
}

// Helper function to initialize a workspace with pm init
func initWorkspace(t *testing.T, pmBinary, workspace, prefix string) {
	t.Helper()

	// Initialize workspace
	output, err := runPM(t, pmBinary, workspace, "init", ".", "--prefix", prefix)
	if err != nil {
		t.Fatalf("pm init failed: %v\nOutput: %s", err, output)
	}

	// Verify .pm directory was created
	pmDir := filepath.Join(workspace, ".pm")
	if _, err := os.Stat(pmDir); os.IsNotExist(err) {
		t.Fatalf("pm init did not create .pm directory")
	}
}

func TestIntegrationWorkflow(t *testing.T) {
	// Build the binary
	pmBinary := buildPMBinary(t)

	// Create test workspace
	workspace := t.TempDir()

	// Step 1: Initialize
	t.Run("init", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "init", ".", "--prefix", "INT")
		if err != nil {
			t.Fatalf("pm init failed: %v\nOutput: %s", err, output)
		}

		// Verify .pm directory was created
		pmDir := filepath.Join(workspace, ".pm")
		if _, err := os.Stat(pmDir); os.IsNotExist(err) {
			t.Error("pm init did not create .pm directory")
		}

		// Verify config files exist
		workflowPath := filepath.Join(pmDir, "config", "workflow.yaml")
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			t.Error("pm init did not create workflow.yaml")
		}

		projectPath := filepath.Join(pmDir, "config", "project.yaml")
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			t.Error("pm init did not create project.yaml")
		}

		// Verify database was created
		dbPath := filepath.Join(pmDir, ".cache.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("pm init did not create .cache.db")
		}
	})

	// Step 2: Create first ticket
	var ticket1ID string
	t.Run("new_first_ticket", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "new", "--type", "task", "First test ticket")
		if err != nil {
			t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
		}

		// Extract ticket ID from output (should be INT-1)
		if strings.Contains(output, "INT-1") {
			ticket1ID = "INT-1"
		} else {
			t.Fatalf("pm new output did not contain expected ticket ID INT-1: %s", output)
		}

		// Verify ticket file exists
		ticketPath := filepath.Join(workspace, ".pm", "tickets", ticket1ID+".md")
		if _, err := os.Stat(ticketPath); os.IsNotExist(err) {
			t.Errorf("pm new did not create ticket file: %s", ticketPath)
		}

		// Verify ticket content
		content, err := os.ReadFile(ticketPath)
		if err != nil {
			t.Fatalf("Failed to read ticket file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "id: INT-1") {
			t.Errorf("Ticket file does not contain correct ID. Content:\n%s", contentStr)
		}
		if !strings.Contains(contentStr, "title: \"First test ticket\"") {
			t.Errorf("Ticket file does not contain correct title. Content:\n%s", contentStr)
		}
		if !strings.Contains(contentStr, "type: task") {
			t.Errorf("Ticket file does not contain correct type. Content:\n%s", contentStr)
		}
		if !strings.Contains(contentStr, "status: backlog") {
			t.Errorf("Ticket file does not contain initial status. Content:\n%s", contentStr)
		}
	})

	// Step 3: Create second ticket
	t.Run("new_second_ticket", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "new", "--type", "bug", "Second test ticket")
		if err != nil {
			t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(output, "INT-2") {
			t.Fatalf("pm new output did not contain expected ticket ID INT-2: %s", output)
		}
	})

	// Step 4: List tickets
	t.Run("list", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "list")
		if err != nil {
			t.Fatalf("pm list failed: %v\nOutput: %s", err, output)
		}

		// Verify both tickets appear in output
		if !strings.Contains(output, "INT-1") {
			t.Error("pm list output does not contain INT-1")
		}
		if !strings.Contains(output, "INT-2") {
			t.Error("pm list output does not contain INT-2")
		}
		if !strings.Contains(output, "First test ticket") {
			t.Error("pm list output does not contain first ticket title")
		}
		if !strings.Contains(output, "Second test ticket") {
			t.Error("pm list output does not contain second ticket title")
		}
	})

	// Step 5: Show ticket
	t.Run("show", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "show", ticket1ID)
		if err != nil {
			t.Fatalf("pm show failed: %v\nOutput: %s", err, output)
		}

		// Verify ticket details are displayed
		if !strings.Contains(output, "INT-1") {
			t.Error("pm show output does not contain ticket ID")
		}
		if !strings.Contains(output, "First test ticket") {
			t.Error("pm show output does not contain ticket title")
		}
	})

	// Step 6: Move ticket
	t.Run("move", func(t *testing.T) {
		// TIMING NOTE: Sleep ensures the file modification time will be in a different second
		// than the previous operation. The cache sync logic uses second-level precision (see
		// ShouldSync() in internal/cache/sync.go), so operations within the same second won't
		// trigger a re-sync. 1100ms guarantees we cross a second boundary.
		time.Sleep(1100 * time.Millisecond)

		output, err := runPM(t, pmBinary, workspace, "move", ticket1ID, "todo")
		if err != nil {
			t.Fatalf("pm move failed: %v\nOutput: %s", err, output)
		}

		// Verify ticket file was updated
		ticketPath := filepath.Join(workspace, ".pm", "tickets", ticket1ID+".md")
		content, err := os.ReadFile(ticketPath)
		if err != nil {
			t.Fatalf("Failed to read ticket file: %v", err)
		}

		if !strings.Contains(string(content), "status: todo") {
			t.Error("pm move did not update ticket status to 'todo'")
		}
	})

	// Step 7: Edit ticket (using --field)
	t.Run("edit_field", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "edit", ticket1ID, "--field", "priority=high")
		if err != nil {
			t.Fatalf("pm edit failed: %v\nOutput: %s", err, output)
		}

		// Verify ticket file was updated
		ticketPath := filepath.Join(workspace, ".pm", "tickets", ticket1ID+".md")
		content, err := os.ReadFile(ticketPath)
		if err != nil {
			t.Fatalf("Failed to read ticket file: %v", err)
		}

		if !strings.Contains(string(content), "priority: high") {
			t.Error("pm edit did not update ticket priority to 'high'")
		}
	})

	// Step 8: List with filter
	t.Run("list_filtered", func(t *testing.T) {
		// TIMING NOTE: Sleep ensures cache detects the file change from the move operation.
		// Without this, the move and list operations may happen in the same second, causing
		// ShouldSync() to return false and the cache to show stale data.
		time.Sleep(1100 * time.Millisecond)

		output, err := runPM(t, pmBinary, workspace, "list", "--status", "todo")
		if err != nil {
			t.Fatalf("pm list --status failed: %v\nOutput: %s", err, output)
		}

		// Should show INT-1 (moved to todo) but not INT-2 (still in backlog)
		if !strings.Contains(output, "INT-1") {
			t.Errorf("pm list --status=todo does not show INT-1\nOutput: %s", output)
		}
		if strings.Contains(output, "INT-2") {
			t.Error("pm list --status=todo incorrectly shows INT-2")
		}
	})

	// Step 9: Test gap handling in ID generation
	t.Run("id_gap_handling", func(t *testing.T) {
		// Delete INT-1 to create a gap
		ticketPath := filepath.Join(workspace, ".pm", "tickets", "INT-1.md")
		if err := os.Remove(ticketPath); err != nil {
			t.Fatalf("Failed to remove ticket: %v", err)
		}

		// Create new ticket - should be INT-3, not INT-1
		output, err := runPM(t, pmBinary, workspace, "new", "Gap test ticket")
		if err != nil {
			t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(output, "INT-3") {
			t.Errorf("pm new did not handle gap correctly, expected INT-3, output: %s", output)
		}
	})
}

func TestIntegrationInitValidation(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()

	// Test init without prefix
	t.Run("init_without_prefix", func(t *testing.T) {
		_, err := runPM(t, pmBinary, workspace, "init", ".")
		if err == nil {
			t.Error("pm init without --prefix should fail, but succeeded")
		}
	})

	// Test init with lowercase prefix (should convert to uppercase)
	t.Run("init_lowercase_prefix", func(t *testing.T) {
		workspace := t.TempDir()
		output, err := runPM(t, pmBinary, workspace, "init", ".", "--prefix", "lower")
		if err != nil {
			t.Fatalf("pm init failed: %v\nOutput: %s", err, output)
		}

		// Verify prefix was uppercased in project.yaml
		projectPath := filepath.Join(workspace, ".pm", "config", "project.yaml")
		content, err := os.ReadFile(projectPath)
		if err != nil {
			t.Fatalf("Failed to read project.yaml: %v", err)
		}

		if !strings.Contains(string(content), "prefix: LOWER") {
			t.Error("pm init did not uppercase prefix in project.yaml")
		}

		// Create a ticket to verify uppercase prefix is used
		output, err = runPM(t, pmBinary, workspace, "new", "Test ticket")
		if err != nil {
			t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(output, "LOWER-1") {
			t.Error("pm new did not use uppercased prefix")
		}
	})
}

// TestHistorySingleChange tests pm history with one status change

// TestShowWithComments tests displaying a ticket with comments
