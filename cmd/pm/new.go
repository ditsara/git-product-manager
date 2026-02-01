package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/segmentio/ksuid"
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

		// For now, assume .pm is in the current directory.
		// A more robust solution will find the .pm directory in parent folders.
		pmPath := ".pm"
		templatePath := filepath.Join(pmPath, "config", "templates", ticketType+".md")

		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			fmt.Printf("Error: template for type '%s' not found at %s\n", ticketType, templatePath)
			os.Exit(1)
		}

		id := "TICKET-" + ksuid.New().String() // Simplified prefix for now
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

		fmt.Printf("✓ Created new ticket: %s\n", ticketPath)

		// TODO: Open in editor
		// TODO: Add to git staging area
	},
}

func init() {
	newCmd.Flags().StringP("type", "t", "story", "Type of the ticket (e.g., story, task, bug, epic)")
	rootCmd.AddCommand(newCmd)
}
