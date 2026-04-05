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

func main() {
	Execute()
}
