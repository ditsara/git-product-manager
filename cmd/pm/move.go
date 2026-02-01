package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var moveCmd = &cobra.Command{
	Use:   "move [id] [status]",
	Short: "Move a ticket to a new status",
	Long:  `Updates the status of a ticket.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ticketID := args[0]
		newStatus := args[1]

		workflow, err := config.LoadWorkflow(filepath.Join(".pm", "config", "workflow.yaml"))
		if err != nil {
			log.Fatalf("Error loading workflow: %v", err)
		}

		if !workflow.IsValidState(newStatus) {
			fmt.Printf("Error: Invalid status '%s'. Valid states are: %v\n", newStatus, workflow.States)
			os.Exit(1)
		}

		ticketPath := findTicketByID(ticketID)

		content, err := os.ReadFile(ticketPath)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}

		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			fmt.Println("Error: Invalid ticket format.")
			os.Exit(1)
		}

		var ticketData map[string]interface{}
		if err := yaml.Unmarshal([]byte(parts[1]), &ticketData); err != nil {
			fmt.Printf("Error parsing YAML: %v\n", err)
			os.Exit(1)
		}

		ticketData["status"] = newStatus
		ticketData["updated_at"] = time.Now().UTC().Format(time.RFC3339)

		newYaml, err := yaml.Marshal(ticketData)
		if err != nil {
			log.Fatalf("Error marshalling YAML: %v", err)
		}

		newContent := fmt.Sprintf("---\n%s---\n%s", string(newYaml), parts[2])
		if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
			log.Fatalf("Error writing updated file: %v", err)
		}

		fmt.Printf("✓ Moved ticket %s to %s\n", ticketID, newStatus)
		// TODO: Auto-commit with message
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
