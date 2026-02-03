package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findTicketByID locates a ticket file by ID with exact match priority.
// It first tries an exact match (ID.md), then falls back to prefix matching.
// If multiple files match the prefix, it returns an error listing all matches.
// Matching is case-insensitive for user convenience.
func findTicketByID(id string) string {
	ticketsPath := ".pm/tickets"

	// Try exact match first (case-insensitive)
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		fmt.Printf("Error reading tickets directory: %v\n", err)
		os.Exit(1)
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
			return filepath.Join(ticketsPath, f.Name())
		}
	}

	// Second pass: find by prefix (case-insensitive), but detect ambiguity
	var matches []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// Compare uppercase versions for case-insensitive matching
		if strings.HasPrefix(strings.ToUpper(f.Name()), normalizedID) {
			matches = append(matches, f.Name())
		}
	}

	if len(matches) == 0 {
		fmt.Printf("Error: ticket '%s' not found.\n", id)
		os.Exit(1)
	}

	if len(matches) > 1 {
		fmt.Printf("Error: ambiguous ID '%s'. Multiple matches found:\n", id)
		for _, match := range matches {
			// Strip .md extension for display
			displayID := strings.TrimSuffix(match, ".md")
			fmt.Printf("  - %s\n", displayID)
		}
		fmt.Println("\nPlease use a more specific ID.")
		os.Exit(1)
	}

	return filepath.Join(ticketsPath, matches[0])
}
