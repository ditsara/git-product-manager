package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tickets",
	Long:  `Lists all tickets from the .pm/tickets directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// For now, assume .pm is in the current directory.
		ticketsPath := ".pm/tickets"
		files, err := os.ReadDir(ticketsPath)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%-20s %-50s %-10s %-15s\n", "ID", "TITLE", "TYPE", "STATUS")
		fmt.Println(strings.Repeat("-", 95))

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				filePath := filepath.Join(ticketsPath, file.Name())
				content, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Error reading file %s: %v", filePath, err)
					continue
				}

				// Simple parsing of YAML front matter
				parts := strings.SplitN(string(content), "---", 3)
				if len(parts) < 3 {
					continue
				}

				var ticket struct {
					ID     string `yaml:"id"`
					Title  string `yaml:"title"`
					Type   string `yaml:"type"`
					Status string `yaml:"status"`
				}

				if err := yaml.Unmarshal([]byte(parts[1]), &ticket); err != nil {
					log.Printf("Error unmarshalling YAML from %s: %v", filePath, err)
					continue
				}

				fmt.Printf("%-20s %-50s %-10s %-15s\n", ticket.ID, ticket.Title, ticket.Type, ticket.Status)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
