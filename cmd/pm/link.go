package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <id> <target-id>",
	Short: "Create a relationship between two tickets",
	Long: `Create a relationship between tickets with optional symmetry.

Supported relationship types:
  depends-on    Source depends on target (bidirectional with blocks)
  blocks        Source blocks target (bidirectional with depends-on)
  related       Loose association between tickets (unidirectional)
  
Default type: related

Examples:
  pm link GPM-5 GPM-10 --type depends-on
  pm link GPM-5 GPM-99 --type related
  pm link GPM-10 GPM-5 --type blocks

Symmetric relationships (depends-on/blocks) automatically update both tickets.
Related links only update the source ticket.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sourceID := args[0]
		targetID := args[1]
		relType, _ := cmd.Flags().GetString("type")
		pmPath := ".pm"

		// Call the relationship helper
		alreadyExists, err := ticket.UpdateRelationshipWithSymmetry(pmPath, sourceID, targetID, relType, true)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Normalize IDs to uppercase for output
		sourceID = strings.ToUpper(sourceID)
		targetID = strings.ToUpper(targetID)

		// If already linked, show that message
		if alreadyExists {
			fmt.Printf("✓ Already linked: %s -> %s (%s)\n", sourceID, targetID, relType)
			return
		}

		// Determine number of files modified
		fileCount := 1
		if relType == "depends-on" || relType == "blocks" {
			fileCount = 2
		}

		if fileCount == 1 {
			fmt.Printf("✓ Linked %s -> %s (%s) [modified %d file]\n", sourceID, targetID, relType, fileCount)
		} else {
			fmt.Printf("✓ Linked %s -> %s (%s) [modified %d files]\n", sourceID, targetID, relType, fileCount)
		}
	},
}

func init() {
	linkCmd.Flags().StringP("type", "T", "related", "Relationship type (depends-on, blocks, related)")

	// Register completion functions for flags
	linkCmd.RegisterFlagCompletionFunc("type", completeRelationshipTypes)

	rootCmd.AddCommand(linkCmd)
}

// completeRelationshipTypes provides completion for relationship types
func completeRelationshipTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"depends-on", "blocks", "related"}, cobra.ShellCompDirectiveNoFileComp
}
