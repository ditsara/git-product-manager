package ticket

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// UpdateRelationshipWithSymmetry manages array-based ticket relationships with automatic symmetry.
// For depends-on/blocks pairs, modifying one ticket automatically updates the inverse in the other.
// For related links, only the source ticket is modified.
//
// Parameters:
//   - sourceID: The ticket being modified (e.g., GPM-5)
//   - targetID: The related ticket (e.g., GPM-10)
//   - relType: Relationship type (depends-on, blocks, related)
//   - add: true to add relationship, false to remove
//
// Returns (alreadyExists, error):
//   - alreadyExists: true if relationship already existed (for add operations)
//   - error: if validation fails or file operations fail
//
// On symmetry operation failure, rolls back the source ticket.
func UpdateRelationshipWithSymmetry(pmPath, sourceID, targetID, relType string, add bool) (bool, error) {
	// Normalize IDs to uppercase (case-insensitive input)
	sourceID = strings.ToUpper(sourceID)
	targetID = strings.ToUpper(targetID)

	// Validate inputs
	if sourceID == targetID {
		return false, fmt.Errorf("cannot link ticket to itself (%s)", sourceID)
	}

	if !isValidRelType(relType) {
		return false, fmt.Errorf("invalid type '%s' - must be depends-on, blocks, or related", relType)
	}

	// Verify both tickets exist before modifying anything
	sourcePath := filepath.Join(pmPath, "tickets", sourceID+".md")
	targetPath := filepath.Join(pmPath, "tickets", targetID+".md")

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return false, fmt.Errorf("source ticket not found: %s", sourceID)
	}
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return false, fmt.Errorf("target ticket not found: %s", targetID)
	}

	// For non-symmetric relationships, just update source ticket
	if relType == "related" {
		return updateRelationshipInFile(sourcePath, targetID, relType, add)
	}

	// For symmetric relationships (depends-on/blocks), we need atomic operations
	// Read both tickets first
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to read source ticket: %w", err)
	}

	targetContent, err := os.ReadFile(targetPath)
	if err != nil {
		return false, fmt.Errorf("failed to read target ticket: %w", err)
	}

	// Parse both tickets
	sourceTicket, err := Parse(sourceContent)
	if err != nil {
		return false, fmt.Errorf("failed to parse source ticket: %w", err)
	}

	targetTicket, err := Parse(targetContent)
	if err != nil {
		return false, fmt.Errorf("failed to parse target ticket: %w", err)
	}

	// Determine inverse relationship type
	inverseType := getInverseType(relType)

	// Check if relationship already exists (for add operations)
	if add {
		if contains(sourceTicket.getArrayForType(relType), targetID) {
			return true, nil // Already exists, return true to indicate "already linked"
		}
	}

	// Modify arrays
	if add {
		sourceTicket.addToArray(relType, targetID)
		targetTicket.addToArray(inverseType, sourceID)
	} else {
		sourceTicket.removeFromArray(relType, targetID)
		targetTicket.removeFromArray(inverseType, sourceID)
	}

	// Normalize to remove any duplicates
	sourceTicket.Normalize()
	targetTicket.Normalize()

	// Update timestamps
	now := time.Now().UTC().Format(time.RFC3339)
	sourceTicket.UpdatedAt = now
	targetTicket.UpdatedAt = now

	// Write source ticket
	if err := writeTicketFile(sourcePath, sourceContent, sourceTicket); err != nil {
		return false, fmt.Errorf("failed to write source ticket: %w", err)
	}

	// Write target ticket - if this fails, we need to roll back source
	if err := writeTicketFile(targetPath, targetContent, targetTicket); err != nil {
		// Rollback: restore original source content
		_ = os.WriteFile(sourcePath, sourceContent, 0644)
		return false, fmt.Errorf("failed to write target ticket (%s), rolled back changes to source: %w", targetID, err)
	}

	return false, nil
}

// isValidRelType checks if the relationship type is valid
func isValidRelType(relType string) bool {
	return relType == "depends-on" || relType == "blocks" || relType == "related"
}

// getInverseType returns the inverse relationship type for symmetric pairs
func getInverseType(relType string) string {
	if relType == "depends-on" {
		return "blocks"
	}
	if relType == "blocks" {
		return "depends-on"
	}
	return relType // related is unidirectional
}

// contains checks if a slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// getArrayForType returns a reference to the appropriate array for the given type
func (t *Ticket) getArrayForType(relType string) []string {
	switch relType {
	case "depends-on":
		return t.DependsOn
	case "blocks":
		return t.Blocks
	case "related":
		return t.Related
	default:
		return []string{}
	}
}

// addToArray adds a value to the specified relationship array
func (t *Ticket) addToArray(relType, value string) {
	switch relType {
	case "depends-on":
		t.DependsOn = append(t.DependsOn, value)
	case "blocks":
		t.Blocks = append(t.Blocks, value)
	case "related":
		t.Related = append(t.Related, value)
	}
}

// removeFromArray removes a value from the specified relationship array
func (t *Ticket) removeFromArray(relType, value string) {
	switch relType {
	case "depends-on":
		t.DependsOn = removeElement(t.DependsOn, value)
	case "blocks":
		t.Blocks = removeElement(t.Blocks, value)
	case "related":
		t.Related = removeElement(t.Related, value)
	}
}

// removeElement removes a value from a slice
func removeElement(slice []string, value string) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}

// updateRelationshipInFile updates a relationship in a single ticket file (unidirectional)
// Returns (alreadyExists, error)
func updateRelationshipInFile(ticketPath, targetID, relType string, add bool) (bool, error) {
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		return false, fmt.Errorf("failed to read ticket: %w", err)
	}

	ticket, err := Parse(content)
	if err != nil {
		return false, fmt.Errorf("failed to parse ticket: %w", err)
	}

	// Check if relationship already exists (for add operations)
	if add {
		if contains(ticket.getArrayForType(relType), targetID) {
			return true, nil // Already exists
		}
		ticket.addToArray(relType, targetID)
	} else {
		ticket.removeFromArray(relType, targetID)
	}

	// Normalize to remove any duplicates
	ticket.Normalize()

	// Update timestamp
	ticket.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	return false, writeTicketFile(ticketPath, content, ticket)
}

// writeTicketFile serializes and writes a ticket to file
// Preserves the original markdown body if present
func writeTicketFile(ticketPath string, originalContent []byte, ticket *Ticket) error {
	// Extract the original body from front matter
	parts := strings.SplitN(string(originalContent), "---", 3)
	var body string
	if len(parts) >= 3 {
		body = parts[2]
	}

	// Marshal YAML
	yamlData, err := yaml.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Reconstruct file with front matter + body
	newContent := fmt.Sprintf("---\n%s---\n%s", string(yamlData), body)

	if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// RemoveRelationshipFromAllFields removes a target ID from all relationship arrays in a ticket
// Used by pm unlink when no type is specified
// Returns (wasNotLinked, error)
func RemoveRelationshipFromAllFields(pmPath, sourceID, targetID string) (bool, error) {
	sourceID = strings.ToUpper(sourceID)
	targetID = strings.ToUpper(targetID)

	if sourceID == targetID {
		return false, fmt.Errorf("cannot unlink ticket from itself (%s)", sourceID)
	}

	sourcePath := filepath.Join(pmPath, "tickets", sourceID+".md")
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return false, fmt.Errorf("ticket not found: %s", sourceID)
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to read ticket: %w", err)
	}

	ticket, err := Parse(content)
	if err != nil {
		return false, fmt.Errorf("failed to parse ticket: %w", err)
	}

	// Remove from all relationship arrays
	found := false
	if contains(ticket.DependsOn, targetID) {
		ticket.DependsOn = removeElement(ticket.DependsOn, targetID)
		found = true
	}
	if contains(ticket.Blocks, targetID) {
		ticket.Blocks = removeElement(ticket.Blocks, targetID)
		found = true
	}
	if contains(ticket.Related, targetID) {
		ticket.Related = removeElement(ticket.Related, targetID)
		found = true
	}

	if !found {
		return true, nil // Not linked, return true to indicate "not linked"
	}

	// Normalize and update timestamp
	ticket.Normalize()
	ticket.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	return false, writeTicketFile(sourcePath, content, ticket)
}
