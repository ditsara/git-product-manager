package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ditsara/git-product-manager/internal/cache"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new Git Product Manager repository",
	Long:  `Initializes a new Git Product Manager repository by creating the .pm directory and default configuration files.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var root string
		if len(args) > 0 {
			root = args[0]
		} else {
			root = "."
		}

		pmPath := filepath.Join(root, ".pm")
		if _, err := os.Stat(pmPath); !os.IsNotExist(err) {
			fmt.Println("Error: .pm directory already exists. Use --force to reinitialize.")
			os.Exit(1)
		}

		fmt.Printf("Initializing Git Product Manager in %s\n", pmPath)

		// Create directories
		dirs := []string{
			"tickets",
			"config/templates",
		}
		for _, dir := range dirs {
			if err := os.MkdirAll(filepath.Join(pmPath, dir), 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}

		// Create config files
		createDefaultWorkflow(pmPath)
		createDefaultLabels(pmPath)
		createDefaultTemplates(pmPath)
		createGitignore(pmPath)

		// Initialize database
		dbPath := filepath.Join(pmPath, ".cache.db")
		// For Stage 1, we assume migrations are in a "migrations" folder relative to the execution path.
		// A more robust solution might be needed for different execution contexts.
		if err := cache.RunMigrations(dbPath, "migrations"); err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Initialized .pm directory")
		fmt.Println("✓ Created default workflow with 4 states")
		fmt.Println("✓ Created default labels")
		fmt.Println("✓ Created 4 ticket templates")
		fmt.Println("\nNext steps:")
		fmt.Println("  pm new \"Your first ticket\"")
		fmt.Println("  pm list")
	},
}

func createDefaultWorkflow(pmPath string) {
	content := `
states:
  - backlog
  - todo
  - in-progress
  - done

initial_state: backlog

state_groups:
  active: [todo, in-progress]
  completed: [done]
  incomplete: [backlog, todo, in-progress]
`
	path := filepath.Join(pmPath, "config", "workflow.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating workflow.yaml: %v\n", err)
		os.Exit(1)
	}
}

func createDefaultLabels(pmPath string) {
	content := `
labels:
  - backend
  - frontend
  - bug
  - feature
  - documentation
  - testing
  - refactor
  - chore
`
	path := filepath.Join(pmPath, "config", "labels.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating labels.yaml: %v\n", err)
		os.Exit(1)
	}
}

func createDefaultTemplates(pmPath string) {
	templates := map[string]string{
		"story.md": `---
id: {{.ID}}
title: "{{.Title}}"
type: story
status: backlog
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: []
assignee: ""
created_at: {{.CreatedAt}}
updated_at: {{.UpdatedAt}}
---

# Description

`,
		"task.md": `---
id: {{.ID}}
title: "{{.Title}}"
type: task
status: backlog
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: []
assignee: ""
created_at: {{.CreatedAt}}
updated_at: {{.UpdatedAt}}
---

# Description

`,
		"bug.md": `---
id: {{.ID}}
title: "{{.Title}}"
type: bug
status: backlog
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: []
assignee: ""
created_at: {{.CreatedAt}}
updated_at: {{.UpdatedAt}}
---

# Description

`,
		"epic.md": `---
id: {{.ID}}
title: "{{.Title}}"
type: epic
status: backlog
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: []
assignee: ""
created_at: {{.CreatedAt}}
updated_at: {{.UpdatedAt}}
---

# Description

`,
	}

	for name, content := range templates {
		path := filepath.Join(pmPath, "config", "templates", name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Printf("Error creating template %s: %v\n", name, err)
			os.Exit(1)
		}
	}
}

func createGitignore(pmPath string) {
	content := `
.cache.db
`
	path := filepath.Join(pmPath, ".gitignore")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating .gitignore: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
}
