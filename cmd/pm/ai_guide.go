package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ditsara/git-product-manager/internal/guide"
	"github.com/spf13/cobra"
)

var aiGuideCmd = &cobra.Command{
	Use:   "guide [section]",
	Short: "Display workflow guidance for LLM assistants",
	Long: `Display GPM workflow guidance as Markdown.

Sections:
  workflow    Step-by-step development process (read this first)
  schema      Ticket YAML field reference
  commands    Command cheat sheet
  principles  Key rules and conventions

With no arguments, outputs all sections — useful for manual one-shot loading:
  pm ai guide > CLAUDE.md

With a section name, outputs only that section:
  pm ai guide workflow`,
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
	aiCmd.AddCommand(aiGuideCmd)
}
