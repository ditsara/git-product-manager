package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var assignCmd = &cobra.Command{
	Use:   "assign <id> <user>",
	Short: "Assign a ticket to a user",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeTicketIDs(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Long: `Quickly assign a ticket to a user without opening an editor.

This is a shorthand for: pm edit <id> --field assignee=<user>

Examples:
  pm assign GPM-123 alice
  pm assign GPM-456 bob@example.com
  pm assign test-1 engineering-team`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ticketID := args[0]
		username := args[1]

		// Validate username is not empty
		if username == "" {
			fmt.Println("Error: Username cannot be empty")
			os.Exit(1)
		}

		// Find the ticket file
		ticketPath := getTicketPath(ticketID)

		// Read the current assignee
		content, err := os.ReadFile(ticketPath)
		if err != nil {
			log.Fatalf("Error reading ticket: %v", err)
		}

		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) != 3 {
			log.Fatal("Invalid ticket format")
		}

		var metadata map[string]interface{}
		if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
			log.Fatal(err)
		}

		// Get current assignee
		currentAssignee := ""
		if assignee, ok := metadata["assignee"]; ok {
			currentAssignee = assignee.(string)
		}

		// Check if already assigned to this user (idempotent)
		if currentAssignee == username {
			fmt.Printf("Already assigned to %s\n", username)
			return
		}

		// Update the assignee field
		metadata["assignee"] = username
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

		fmt.Printf("✓ Assigned %s to %s\n", ticketID, username)
	},
}

func init() {
	rootCmd.AddCommand(assignCmd)
}
