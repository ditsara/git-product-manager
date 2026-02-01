package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/ditsara/git-product-manager/internal/config"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create a new ticket",
	Long:  `Creates a new ticket file from a template.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		ticketType, _ := cmd.Flags().GetString("type")

		pmPath := ".pm"
		
		// Load project config to get prefix
		project, err := config.LoadProject(pmPath)
		if err != nil {
			fmt.Printf("Error loading project config: %v\n", err)
			fmt.Println("Make sure you've run 'pm init --prefix YOUR_PREFIX' first")
			os.Exit(1)
		}

		templatePath := filepath.Join(pmPath, "config", "templates", ticketType+".md")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			fmt.Printf("Error: template for type '%s' not found at %s\n", ticketType, templatePath)
			os.Exit(1)
		}

		// Generate next sequential ID
		nextNum := getNextTicketNumber(pmPath, project.Prefix)
		id := fmt.Sprintf("%s-%d", project.Prefix, nextNum)
		now := time.Now().UTC().Format(time.RFC3339)

		data := struct {
			ID        string
			Title     string
			CreatedAt string
			UpdatedAt string
		}{
			ID:        id,
			Title:     title,
			CreatedAt: now,
			UpdatedAt: now,
		}

		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Printf("Error parsing template: %v\n", err)
			os.Exit(1)
		}

		ticketPath := filepath.Join(pmPath, "tickets", id+".md")
		file, err := os.Create(ticketPath)
		if err != nil {
			fmt.Printf("Error creating ticket file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		if err := tmpl.Execute(file, data); err != nil {
			fmt.Printf("Error executing template: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Created new ticket: %s\n", id)

		// TODO: Open in editor
		// TODO: Add to git staging area
	},
}

// getNextTicketNumber scans the tickets directory and returns the next sequential number
func getNextTicketNumber(pmPath string, prefix string) int {
	ticketsPath := filepath.Join(pmPath, "tickets")
	files, err := os.ReadDir(ticketsPath)
	if err != nil {
		// If directory doesn't exist or can't be read, start at 1
		return 1
	}

	maxNum := 0
	// Pattern: PREFIX-123.md
	pattern := regexp.MustCompile(fmt.Sprintf(`^%s-(\d+)\.md$`, regexp.QuoteMeta(prefix)))

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		matches := pattern.FindStringSubmatch(file.Name())
		if len(matches) == 2 {
			num, err := strconv.Atoi(matches[1])
			if err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	return maxNum + 1
}

func init() {
	newCmd.Flags().StringP("type", "t", "story", "Type of the ticket (e.g., story, task, bug, epic)")
	rootCmd.AddCommand(newCmd)
}
