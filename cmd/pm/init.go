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
			fmt.Printf("GPM already initialised at %s — regenerating missing files...\n", pmPath)
		} else {
			fmt.Printf("Initializing Git Product Manager in %s\n", pmPath)
		}

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
		fmt.Println("\nNext steps:")
		fmt.Println("  pm new \"Your first ticket\"")
		fmt.Println("  pm list")
		fmt.Println("  pm ai guide      # read workflow guidance")
	},
}

func createDefaultWorkflow(pmPath string) {
	path := filepath.Join(pmPath, "config", "workflow.yaml")
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  ✓ %s (exists, skipping)\n", path)
		return
	}
	content, err := templateFS.ReadFile("templates/workflow.yaml")
	if err != nil {
		fmt.Printf("Error reading embedded workflow.yaml: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Printf("Error creating workflow.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Created %s\n", path)
}

func createDefaultLabels(pmPath string) {
	path := filepath.Join(pmPath, "config", "labels.yaml")
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  ✓ %s (exists, skipping)\n", path)
		return
	}
	content, err := templateFS.ReadFile("templates/labels.yaml")
	if err != nil {
		fmt.Printf("Error reading embedded labels.yaml: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Printf("Error creating labels.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Created %s\n", path)
}

func createDefaultTemplates(pmPath string) {
	templateNames := []string{"story.md", "task.md", "bug.md", "epic.md", "milestone.md"}

	for _, name := range templateNames {
		path := filepath.Join(pmPath, "config", "templates", name)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✓ %s (exists, skipping)\n", path)
			continue
		}
		content, err := templateFS.ReadFile(filepath.Join("templates", name))
		if err != nil {
			fmt.Printf("Error reading embedded template %s: %v\n", name, err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			fmt.Printf("Error creating template %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Created %s\n", path)
	}
}

func createGitignore(pmPath string) {
	path := filepath.Join(pmPath, ".gitignore")
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  ✓ %s (exists, skipping)\n", path)
		return
	}
	content := "\n.cache.db\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating .gitignore: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Created %s\n", path)
}

func createProjectConfig(pmPath string, prefix string) {
	path := filepath.Join(pmPath, "config", "project.yaml")
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  ✓ %s (exists, skipping)\n", path)
		return
	}
	project := &config.Project{
		Prefix: prefix,
	}
	if err := config.SaveProject(pmPath, project); err != nil {
		fmt.Printf("Error creating project.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Created %s (prefix: %s)\n", path, prefix)
}

func init() {
	initCmd.Flags().StringP("prefix", "p", "", "Ticket prefix (e.g., MYPROJECT) - required")
	initCmd.MarkFlagRequired("prefix")
	rootCmd.AddCommand(initCmd)
}
