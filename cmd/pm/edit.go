package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var editCmd = &cobra.Command{
	Use:               "edit [id]",
	Short:             "Edit a ticket",
	ValidArgsFunction: completeTicketIDs,
	Long: `Opens a ticket in the default editor or updates a specific field.

Examples:
  # Open ticket in editor
  pm edit GPM-1

  # Update a single field
  pm edit GPM-1 --field assignee=alice

  # Replace the ticket description/body
  pm edit GPM-1 --description "New body text here"

  # Update array field (replaces all values)
  pm edit GPM-1 --field labels=bug,critical,p1

  # Clear an array field
  pm edit GPM-1 --field labels=

  # Update integer field
  pm edit GPM-1 --field points=5

  # Update enum field
  pm edit GPM-1 --field priority=high

  # Update a field and description together
  pm edit GPM-1 --field priority=high --description "Revised body"

Note: Array fields use comma (,) as delimiter. Values are trimmed.
Array updates REPLACE existing values, they do not append.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ticketID := args[0]
		ticketPath := getTicketPath(ticketID)

		field, _ := cmd.Flags().GetString("field")
		description, _ := cmd.Flags().GetString("description")
		hasField := field != ""
		hasDescription := cmd.Flags().Changed("description")

		if hasField || hasDescription {
			// Parse field=value
			if hasField {
				parts := strings.SplitN(field, "=", 2)
				if len(parts) != 2 {
					fmt.Println("Error: --field must be in format field=value")
					os.Exit(1)
				}
				fieldName, fieldValue := parts[0], parts[1]
				// Validate milestone IDs exist before writing
				if fieldName == "milestones" && fieldValue != "" {
					pmPath := ".pm"
					milestonesDir := filepath.Join(pmPath, "milestones")
					for _, mid := range strings.Split(fieldValue, ",") {
						mid = strings.TrimSpace(mid)
						if mid == "" {
							continue
						}
						milePath := filepath.Join(milestonesDir, mid+".md")
						if _, err := os.Stat(milePath); os.IsNotExist(err) {
							fmt.Fprintf(os.Stderr, "Error: milestone not found: %q\nRun `pm milestone list` to see available milestones.\n", mid)
							os.Exit(1)
						}
					}
				}

				updateTicketField(ticketPath, fieldName, fieldValue)
			}

			if hasDescription {
				updateTicketDescription(ticketPath, description)
			}

			switch {
			case hasField && hasDescription:
				fmt.Printf("✓ Updated fields and description for %s\n", ticketID)
			case hasField:
				fieldName := strings.SplitN(field, "=", 2)[0]
				fmt.Printf("✓ Updated %s for %s\n", fieldName, ticketID)
			default:
				fmt.Printf("✓ Updated description for %s\n", ticketID)
			}
			return
		}

		// Open in editor
		editor := getEditor()
		c := exec.Command(editor, ticketPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			log.Fatalf("Error running editor: %v", err)
		}

		// Read the file to update the 'updated_at' timestamp
		content, err := os.ReadFile(ticketPath)
		if err != nil {
			log.Fatalf("Error reading file after edit: %v", err)
		}

		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			fmt.Println("Error: Invalid ticket format after edit.")
			os.Exit(1)
		}

		var ticketData map[string]interface{}
		if err := yaml.Unmarshal([]byte(parts[1]), &ticketData); err != nil {
			fmt.Printf("Error parsing YAML after edit: %v\n", err)
			// TODO: Offer to re-edit or discard
			os.Exit(1)
		}

		ticketData["updated_at"] = time.Now().UTC().Format(time.RFC3339)

		newYaml, err := yaml.Marshal(ticketData)
		if err != nil {
			log.Fatalf("Error marshalling YAML: %v", err)
		}

		newContent := fmt.Sprintf("---\n%s---\n%s", string(newYaml), parts[2])
		if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
			log.Fatalf("Error writing updated file: %v", err)
		}

		fmt.Printf("✓ Updated ticket: %s\n", ticketPath)
		// TODO: Add to git staging area
	},
}

func getEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	// Check for common editors in PATH
	for _, editor := range []string{"editor", "nano", "vi"} {
		if path, err := exec.LookPath(editor); err == nil {
			return path
		}
	}
	return "vi" // POSIX fallback
}

func init() {
	editCmd.Flags().String("field", "", "Update a specific field (format: field=value)")
	editCmd.Flags().String("description", "", "Replace the ticket body/description")
	rootCmd.AddCommand(editCmd)
}

// updateTicketField updates a specific field in a ticket
func updateTicketField(ticketPath, field, value string) {
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) != 3 {
		log.Fatal("Invalid ticket format")
	}

	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		log.Fatal(err)
	}

	// Parse the value according to field type
	parsedValue, err := ticket.ParseFieldValue(field, value)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Update the field with the correctly typed value
	metadata[field] = parsedValue
	metadata["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	newYAML, err := yaml.Marshal(metadata)
	if err != nil {
		log.Fatal(err)
	}

	// Reconstruct the file
	newContent := "---\n" + string(newYAML) + "---" + parts[2]
	if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
		log.Fatal(err)
	}
}

// updateTicketDescription replaces the markdown body of a ticket.
func updateTicketDescription(ticketPath, description string) {
	content, err := os.ReadFile(ticketPath)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) != 3 {
		log.Fatal("Invalid ticket format")
	}

	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		log.Fatal(err)
	}

	metadata["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	newYAML, err := yaml.Marshal(metadata)
	if err != nil {
		log.Fatal(err)
	}

	newContent := fmt.Sprintf("---\n%s---\n", string(newYAML))
	if description != "" {
		newContent += "\n" + description
	}

	if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
		log.Fatal(err)
	}
}
