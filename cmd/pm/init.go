package main

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/cache"
	"github.com/ditsara/git-product-manager/internal/config"
	"github.com/ditsara/git-product-manager/internal/guide"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

//go:embed all:templates
var templateFS embed.FS

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new Git Product Manager repository",
	Long:  `Initializes a new Git Product Manager repository by creating the .pm directory and default configuration files.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate prefix
		prefix, _ := cmd.Flags().GetString("prefix")
		if prefix == "" {
			fmt.Println("Error: --prefix is required. Please specify a ticket prefix (e.g., --prefix MYPROJECT)")
			fmt.Println("Example: pm init --prefix myproject")
			os.Exit(1)
		}
		// Convert to uppercase
		prefix = strings.ToUpper(prefix)

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
			"milestones",
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
		createProjectConfig(pmPath, prefix)
		createWorkflowGuide(pmPath)
		createAgentsFile(pmPath)

		// Initialize database
		dbPath := filepath.Join(pmPath, ".cache.db")

		if err := cache.RunMigrations(dbPath); err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			os.Exit(1)
		}

		// Verify database is accessible
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			fmt.Printf("Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			fmt.Printf("Error verifying database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Initialized .pm directory")
		fmt.Println("✓ Created default workflow with 4 states")
		fmt.Println("✓ Created default labels")
		fmt.Println("✓ Created 5 ticket templates")
		fmt.Println("✓ Created workflow guide")
		fmt.Printf("✓ Project prefix set to: %s\n", prefix)
		fmt.Println("\nNext steps:")
		fmt.Println("  pm new \"Your first ticket\"")
		fmt.Println("  pm list")
		fmt.Println("  pm guide > CLAUDE.md   # generate LLM guidance file")
	},
}

func createDefaultWorkflow(pmPath string) {
	// Read workflow from embedded filesystem
	content, err := templateFS.ReadFile("templates/workflow.yaml")
	if err != nil {
		fmt.Printf("Error reading embedded workflow.yaml: %v\n", err)
		os.Exit(1)
	}

	path := filepath.Join(pmPath, "config", "workflow.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Printf("Error creating workflow.yaml: %v\n", err)
		os.Exit(1)
	}
}

func createDefaultLabels(pmPath string) {
	// Read labels from embedded filesystem
	content, err := templateFS.ReadFile("templates/labels.yaml")
	if err != nil {
		fmt.Printf("Error reading embedded labels.yaml: %v\n", err)
		os.Exit(1)
	}

	path := filepath.Join(pmPath, "config", "labels.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Printf("Error creating labels.yaml: %v\n", err)
		os.Exit(1)
	}
}

func createDefaultTemplates(pmPath string) {
	templateNames := []string{"story.md", "task.md", "bug.md", "epic.md", "milestone.md"}

	for _, name := range templateNames {
		// Read template from embedded filesystem
		content, err := templateFS.ReadFile(filepath.Join("templates", name))
		if err != nil {
			fmt.Printf("Error reading embedded template %s: %v\n", name, err)
			os.Exit(1)
		}

		// Write to .pm/config/templates/
		path := filepath.Join(pmPath, "config", "templates", name)
		if err := os.WriteFile(path, content, 0644); err != nil {
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

func createProjectConfig(pmPath string, prefix string) {
	project := &config.Project{
		Prefix: prefix,
	}
	if err := config.SaveProject(pmPath, project); err != nil {
		fmt.Printf("Error creating project.yaml: %v\n", err)
		os.Exit(1)
	}
}

func createWorkflowGuide(pmPath string) {
	sections := guide.SectionNames()
	lines := []string{
		"# GPM Workflow Guide",
		"",
		"This project uses Git Product Manager (GPM).",
		"For current guidance, run:",
		"",
	}
	for _, s := range sections {
		lines = append(lines, "  pm guide "+s)
	}
	lines = append(lines,
		"",
		"Or generate a complete guidance file for your LLM:",
		"",
		"  pm guide > CLAUDE.md",
		"",
	)
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	path := filepath.Join(pmPath, "config", "WORKFLOW_GUIDE.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating WORKFLOW_GUIDE.md: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	initCmd.Flags().StringP("prefix", "p", "", "Ticket prefix (e.g., MYPROJECT) - required")
	initCmd.MarkFlagRequired("prefix")
	rootCmd.AddCommand(initCmd)
}
