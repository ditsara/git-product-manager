package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pm",
	Short: "Git Product Manager: project management tool for git repos.",
	Long:  `Git Product Manager: a fast and flexible project management tool for git repositories.`,
	Example: `  # Start a new project
  pm init . --prefix MYPROJECT

  # Create tickets
  pm new "Fix login bug"
  pm new "Auth overhaul" --type epic

  # Browse and triage
  pm list
  pm list --active
  pm list --parent GPM-1
  pm show GPM-42

  # Move work forward
  pm move GPM-42 in-progress
  pm edit GPM-42 --field priority=high
  pm edit GPM-42 --description "Refined implementation notes"
  pm edit GPM-42 --field labels=bug,auth
  pm assign GPM-42 alice

  # Collaborate
  pm comment GPM-42 -m "Blocked on infra"
  pm link GPM-42 --depends-on GPM-10
  pm blocked

  # Milestones
  pm milestone create "v1.0 Release" --due 2026-06-01
  pm milestone list --progress

  # AI / LLM integration
  pm ai init
  pm ai guide
  pm ai guide workflow`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help when no subcommand is provided
		cmd.Help()
		fmt.Printf("\n\nBuilt in BKK with ❤️  ☮️  🤖  😡\n")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
}

func main() {
	Execute()
}
