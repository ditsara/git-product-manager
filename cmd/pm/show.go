package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show <id>",
	Short:   "Show a single ticket",
	Long:    `Shows the details of a single ticket.`,
	Args:    cobra.ExactArgs(1),
	Example: "  pm show PROJ-123",
	Run: func(cmd *cobra.Command, args []string) {
		ticketID := args[0]
		ticketPath := findTicketByID(ticketID)

		content, err := os.ReadFile(ticketPath)
		if err != nil {
			log.Fatalf("Error reading ticket file: %v", err)
		}

		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			fmt.Println("Error: Invalid ticket format. Missing YAML front matter.")
			os.Exit(1)
		}

		fmt.Println("---")
		fmt.Println(strings.TrimSpace(parts[1]))
		fmt.Println("---")
		fmt.Println(strings.TrimSpace(parts[2]))

		// TODO: Render markdown
		// TODO: Show comments (Stage 2)
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
