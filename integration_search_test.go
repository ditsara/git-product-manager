package main

import (
	"strings"
	"testing"
)

func TestIntegrationSearch(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initWorkspace(t, pmBinary, workspace, "SRCH")

	// Create tickets with distinct content
	runPM(t, pmBinary, workspace, "new", "Authentication overhaul", "--type", "epic")
	runPM(t, pmBinary, workspace, "new", "Fix login bug", "--type", "bug")
	runPM(t, pmBinary, workspace, "new", "Refactor database layer", "--type", "task")
	runPM(t, pmBinary, workspace, "edit", "SRCH-3", "--description", "This ticket covers authentication-related database refactoring")
	runPM(t, pmBinary, workspace, "move", "SRCH-2", "done")

	t.Run("finds_by_title", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "search", "login")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "SRCH-2") {
			t.Errorf("expected SRCH-2 in results, got:\n%s", output)
		}
		if !strings.Contains(output, "Fix login bug") {
			t.Errorf("expected title in results, got:\n%s", output)
		}
	})

	t.Run("finds_by_body", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "search", "database refactoring")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "SRCH-3") {
			t.Errorf("expected SRCH-3 in results, got:\n%s", output)
		}
		// snippet appears in the MATCH column (no "Match:" label in table format)
		if !strings.Contains(output, "database refactoring") {
			t.Errorf("expected snippet text in results, got:\n%s", output)
		}
	})

	t.Run("relevance_ordering_title_before_body", func(t *testing.T) {
		// "authentication" matches SRCH-1 title and SRCH-3 body
		output, err := runPM(t, pmBinary, workspace, "search", "authentication")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		idx1 := strings.Index(output, "SRCH-1")
		idx3 := strings.Index(output, "SRCH-3")
		if idx1 < 0 || idx3 < 0 {
			t.Fatalf("expected both SRCH-1 and SRCH-3 in results, got:\n%s", output)
		}
		if idx1 > idx3 {
			t.Errorf("expected title match (SRCH-1) before body match (SRCH-3), got:\n%s", output)
		}
	})

	t.Run("no_results", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "search", "xyzzy-not-found")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "No results") {
			t.Errorf("expected no results message, got:\n%s", output)
		}
	})

	t.Run("status_filter_active", func(t *testing.T) {
		// SRCH-2 is done; active filter should exclude it
		output, err := runPM(t, pmBinary, workspace, "search", "login", "--active")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		if strings.Contains(output, "SRCH-2") {
			t.Errorf("expected SRCH-2 (done) excluded by --active, got:\n%s", output)
		}
	})

	t.Run("status_filter_completed", func(t *testing.T) {
		// Only done tickets
		output, err := runPM(t, pmBinary, workspace, "search", "login", "--completed")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "SRCH-2") {
			t.Errorf("expected SRCH-2 in --completed results, got:\n%s", output)
		}
	})

	t.Run("status_flag", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "search", "authentication", "--status", "backlog")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "SRCH-1") {
			t.Errorf("expected SRCH-1 in --status backlog results, got:\n%s", output)
		}
	})

	t.Run("result_count_in_header", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "search", "authentication")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "2 matches") {
			t.Errorf("expected '2 matches' in header, got:\n%s", output)
		}
	})

	t.Run("table_headers", func(t *testing.T) {
		output, err := runPM(t, pmBinary, workspace, "search", "authentication")
		if err != nil {
			t.Fatalf("pm search failed: %v\nOutput: %s", err, output)
		}
		for _, header := range []string{"ID", "TITLE", "MATCH", "TYPE", "STATUS"} {
			if !strings.Contains(output, header) {
				t.Errorf("expected column header %q in output, got:\n%s", header, output)
			}
		}
	})
}
