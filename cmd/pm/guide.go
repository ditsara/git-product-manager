package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ditsara/git-product-manager/internal/guide"
	"github.com/spf13/cobra"
)

var guideCmd = &cobra.Command{
	Use:   "guide [section]",
	Short: "Display workflow guidance for LLM assistants",
	Long: `Display GPM workflow guidance as Markdown.

With no arguments, outputs the full guide — suitable for piping to a context file:

  pm guide > CLAUDE.md
  pm guide > AGENTS.md

With a section name, outputs only that section:

  pm guide workflow    # Development workflow process
  pm guide schema      # Ticket YAML schema reference
  pm guide commands    # Commands cheat sheet
  pm guide principles  # Key principles (including git commit rules)`,
	ValidArgsFunction: completeGuideSections,
	Args:              cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			content, err := guide.Full()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(content)
			return
		}

		section := strings.ToLower(args[0])
		content, err := guide.Section(section)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(content)
	},
}

func init() {
	rootCmd.AddCommand(guideCmd)
}
