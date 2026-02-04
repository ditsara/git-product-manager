package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
)

var showFlags struct {
	noComments bool
}

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

		// Extract actual ticket ID from path for comment directory
		actualTicketID := strings.TrimSuffix(filepath.Base(ticketPath), ".md")

		// Display comments if not suppressed
		if !showFlags.noComments {
			displayComments(actualTicketID)
		}

		// TODO: Render markdown
	},
}

func init() {
	showCmd.Flags().BoolVarP(&showFlags.noComments, "no-comments", "", false, "Hide comments")
	rootCmd.AddCommand(showCmd)
}

// displayComments reads and displays all comments for a ticket
func displayComments(ticketID string) {
	comments, err := ticket.ListCommentsForTicket(ticketID, ".pm")
	if err != nil {
		// Silently skip comments on error
		return
	}

	// If no comments, don't show comments section
	if len(comments) == 0 {
		return
	}

	// Display separator and comments header
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Comments (%d):\n", len(comments))
	fmt.Println()

	// Display each comment
	for i, comment := range comments {
		// Format timestamp in readable UTC format
		timestamp := comment.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC")
		fmt.Printf("@%s (%s)\n", comment.Author, timestamp)
		fmt.Println(comment.Body)

		// Add blank line between comments (but not after the last one)
		if i < len(comments)-1 {
			fmt.Println()
		}
	}
}
