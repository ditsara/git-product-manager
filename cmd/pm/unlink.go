package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <id> <target-id>",
	Short: "Remove a relationship between two tickets",
	Long: `Remove a relationship between tickets.

Without --type flag, removes the target ID from all relationship arrays.
With --type flag, removes only from the specified relationship type.

Supported relationship types:
  depends-on    Source depends on target (removes bidirectional pair)
  blocks        Source blocks target (removes bidirectional pair)
  related       Loose association between tickets

Examples:
  pm unlink GPM-5 GPM-10 --type depends-on
  pm unlink GPM-5 GPM-99 --type related
  pm unlink GPM-5 GPM-99    # Removes from all relationship types`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sourceID := args[0]
		targetID := args[1]
		relType, _ := cmd.Flags().GetString("type")
		pmPath := ".pm"

		// Normalize IDs to uppercase for output
		sourceID = strings.ToUpper(sourceID)
		targetID = strings.ToUpper(targetID)

		var alreadyRemoved bool
		var err error
		var fileCount int

		if relType == "" {
			// Remove from all relationship arrays (unidirectional)
			alreadyRemoved, err = ticket.RemoveRelationshipFromAllFields(pmPath, sourceID, targetID)
			fileCount = 1
		} else {
			// Remove specific relationship type with symmetry
			alreadyRemoved, err = ticket.UpdateRelationshipWithSymmetry(pmPath, sourceID, targetID, relType, false)
			fileCount = 1
			if relType == "depends-on" || relType == "blocks" {
				fileCount = 2
			}
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Output message
		if alreadyRemoved {
			if relType == "" {
				fmt.Printf("✓ Not linked: %s -> %s\n", sourceID, targetID)
			} else {
				fmt.Printf("✓ Not linked: %s -> %s (%s)\n", sourceID, targetID, relType)
			}
		} else {
			if relType == "" {
				fmt.Printf("✓ Removed %s from all relationships in %s [modified %d file]\n", targetID, sourceID, fileCount)
			} else {
				if fileCount == 1 {
					fmt.Printf("✓ Unlinked %s -> %s (%s) [modified %d file]\n", sourceID, targetID, relType, fileCount)
				} else {
					fmt.Printf("✓ Unlinked %s -> %s (%s) [modified %d files]\n", sourceID, targetID, relType, fileCount)
				}
			}
		}
	},
}

func init() {
	unlinkCmd.Flags().StringP("type", "T", "", "Relationship type to remove (depends-on, blocks, related). If empty, removes from all types.")

	// Register completion functions for flags
	unlinkCmd.RegisterFlagCompletionFunc("type", completeRelationshipTypesOptional)

	rootCmd.AddCommand(unlinkCmd)
}

// completeRelationshipTypesOptional provides completion for relationship types (optional)
func completeRelationshipTypesOptional(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"depends-on", "blocks", "related"}, cobra.ShellCompDirectiveNoFileComp
}
