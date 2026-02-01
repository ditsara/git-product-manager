package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit a ticket",
	Long:  `Opens a ticket in the default editor or updates a specific field.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ticketID := args[0]
		ticketPath := findTicketByID(ticketID)
		
		// Check if --field flag is used
		field, _ := cmd.Flags().GetString("field")
		if field != "" {
			// Parse field=value
			parts := strings.SplitN(field, "=", 2)
			if len(parts) != 2 {
				fmt.Println("Error: --field must be in format field=value")
				os.Exit(1)
			}
			updateTicketField(ticketPath, parts[0], parts[1])
			fmt.Printf("✓ Updated %s for %s\n", parts[0], ticketID)
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

func findTicketByID(idPrefix string) string {
	ticketsPath := ".pm/tickets"
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		log.Fatalf("Error reading tickets directory: %v", err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), idPrefix) {
			return filepath.Join(ticketsPath, f.Name())
		}
	}

	fmt.Printf("Error: ticket with ID starting with '%s' not found.\n", idPrefix)
	os.Exit(1)
	return ""
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

	// Update the field
	metadata[field] = value
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
