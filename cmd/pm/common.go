package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// getEditor returns the editor to use for interactive editing.
// Priority:
//  1. $VISUAL environment variable (cross-platform)
//  2. $EDITOR environment variable (cross-platform)
//  3. Platform-specific fallback chain:
//     - Windows: code.cmd → notepad++.exe → notepad.exe (always present)
//     - Unix/macOS: editor (alternatives) → nano → vi (POSIX)
func getEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	if runtime.GOOS == "windows" {
		for _, editor := range []string{"code.cmd", "notepad++.exe", "notepad.exe"} {
			if path, err := exec.LookPath(editor); err == nil {
				return path
			}
		}
		return "notepad.exe" // Always present on Windows
	}

	// Unix/Linux/macOS fallback chain
	for _, editor := range []string{"editor", "nano", "vi"} {
		if path, err := exec.LookPath(editor); err == nil {
			return path
		}
	}
	return "vi" // POSIX fallback
}

// noColor is set by the --no-color persistent flag.
var noColor bool

// colorEnabled returns true unless --no-color was passed or the NO_COLOR
// environment variable is set (per https://no-color.org/).
func colorEnabled() bool {
	if noColor {
		return false
	}
	_, set := os.LookupEnv("NO_COLOR")
	return !set
}

// resolveTicketID locates a ticket by ID and returns the normalized ID string.
// Returns empty string if not found.
// Matching is case-insensitive for user convenience.
// Handles exact matches first, then prefix matching.
func resolveTicketID(id string) string {
	if id == "" {
		return ""
	}

	ticketsPath := ".pm/tickets"

	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		// Return empty string if directory doesn't exist
		return ""
	}

	// Normalize search ID to uppercase for comparison
	normalizedID := strings.ToUpper(id)
	normalizedIDWithExt := normalizedID + ".MD"

	// First pass: look for exact match (case-insensitive)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.ToUpper(f.Name()) == normalizedIDWithExt {
			// Return the ticket ID without .md extension
			return strings.TrimSuffix(f.Name(), ".md")
		}
	}

	// Second pass: find by prefix (case-insensitive)
	var matches []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(f.Name()), normalizedID) {
			matches = append(matches, f.Name())
		}
	}

	if len(matches) == 0 {
		return ""
	}

	if len(matches) > 1 {
		// Multiple matches - return empty string
		// The caller can decide how to handle this
		return ""
	}

	// Single match found
	return strings.TrimSuffix(matches[0], ".md")
}

// getTicketPath locates a ticket file by ID with exact match priority.
// It first tries an exact match (ID.md), then falls back to prefix matching.
// If multiple files match the prefix, it returns an error listing all matches.
// Matching is case-insensitive for user convenience.
func getTicketPath(id string) string {
	ticketsPath := ".pm/tickets"

	// Use resolveTicketID for the core lookup logic
	foundID := resolveTicketID(id)
	if foundID != "" {
		return filepath.Join(ticketsPath, foundID+".md")
	}

	// If not found with exact/prefix match, check if it was ambiguous or missing
	if id == "" {
		fmt.Printf("Error: ticket '%s' not found.\n", id)
		os.Exit(1)
	}

	// Try to detect ambiguous matches by doing a prefix search
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		fmt.Printf("Error reading tickets directory: %v\n", err)
		os.Exit(1)
	}

	normalizedID := strings.ToUpper(id)
	var matches []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(f.Name()), normalizedID) {
			matches = append(matches, f.Name())
		}
	}

	if len(matches) > 1 {
		fmt.Printf("Error: ambiguous ID '%s'. Multiple matches found:\n", id)
		for _, match := range matches {
			displayID := strings.TrimSuffix(match, ".md")
			fmt.Printf("  - %s\n", displayID)
		}
		fmt.Println("\nPlease use a more specific ID.")
		os.Exit(1)
	}

	fmt.Printf("Error: ticket '%s' not found.\n", id)
	os.Exit(1)
	return ""
}
