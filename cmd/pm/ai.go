package main

import "github.com/spf13/cobra"

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Commands for AI/LLM assistant integration",
}

func init() {
	rootCmd.AddCommand(aiCmd)
}
