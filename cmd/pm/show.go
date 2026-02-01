package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show a single ticket",
	Long:  `Shows the details of a single ticket.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ticketID := args[0]
		// This is a simplification. A real implementation would search for the file.
		ticketPath := filepath.Join(".pm", "tickets", ticketID+".md")

		if _, err := os.Stat(ticketPath); os.IsNotExist(err) {
			// Let's try to find the ticket by prefix
			files, err := os.ReadDir(filepath.Join(".pm", "tickets"))
			if err != nil {
				log.Fatalf("Error reading tickets directory: %v", err)
			}
			found := false
			for _, f := range files {
				if strings.HasPrefix(f.Name(), ticketID) {
					ticketPath = filepath.Join(".pm", "tickets", f.Name())
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Error: ticket with ID starting with '%s' not found.\n", ticketID)
				os.Exit(1)
			}
		}

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
