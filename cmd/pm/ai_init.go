package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const agentsMDContent = `# GPM — Git Product Manager

This project uses GPM for ticket management.

Do not create files or edit YAML front matter manually -- use the ` + "`pm`" + ` CLI.
Ticket, milestone, and comment markdown content can be edited directly.
After editing ticket content directly, run ` + "`pm edit <id> --touch`" + ` to update the timestamp.

Get started:
  pm --help                   # list all commands
  pm ai guide --help          # see available guide sections
  pm ai guide workflow        # read the development workflow
  pm list                     # show open tickets
  pm show <id>                # read a ticket
`

// pointerBlock is appended to tool config files by pm ai init.
const pointerBlock = `
# GPM
# See .pm/AGENTS.md for project management instructions.
`

// aiToolTargets maps tool names to their config file paths (relative to project root).
var aiToolTargets = map[string]string{
	"claude":  "CLAUDE.md",
	"copilot": ".github/copilot-instructions.md",
	"cursor":  ".cursor/rules/gpm.mdc",
	"aider":   "CONVENTIONS.md",
}

var aiToolOrder = []string{"claude", "copilot", "cursor", "aider"}

// createAgentsFile writes .pm/AGENTS.md if it does not already exist.
func createAgentsFile(pmPath string) {
	path := filepath.Join(pmPath, "AGENTS.md")
	if _, err := os.Stat(path); err == nil {
		return // already exists
	}
	if err := os.WriteFile(path, []byte(agentsMDContent), 0644); err != nil {
		fmt.Printf("Error creating AGENTS.md: %v\n", err)
		os.Exit(1)
	}
}

var aiInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Append a GPM pointer to LLM tool config files",
	Long: `Append a short pointer to .pm/AGENTS.md in your LLM tool's config file.

Existing file content is never modified — only appended to. Running this
command multiple times is safe (idempotent).

Supported tools:
  claude    CLAUDE.md
  copilot   .github/copilot-instructions.md
  cursor    .cursor/rules/gpm.mdc
  aider     CONVENTIONS.md

Examples:
  pm ai init               # append to all tool config files
  pm ai init --for claude  # append only to CLAUDE.md`,
	Run: func(cmd *cobra.Command, args []string) {
		forTool, _ := cmd.Flags().GetString("for")
		forTool = strings.ToLower(forTool)

		var targets []string
		if forTool == "" || forTool == "all" {
			targets = aiToolOrder
		} else {
			if _, ok := aiToolTargets[forTool]; !ok {
				fmt.Fprintf(os.Stderr, "Error: unknown tool %q. Valid options: %s, all\n",
					forTool, strings.Join(aiToolOrder, ", "))
				os.Exit(1)
			}
			targets = []string{forTool}
		}

		for _, tool := range targets {
			path := aiToolTargets[tool]
			appendPointer(path)
		}
	},
}

// appendPointer appends pointerBlock to path if not already present.
// Creates the file (and any parent directories) if it does not exist.
func appendPointer(path string) {
	// Read existing content (empty string if file doesn't exist)
	existing := ""
	data, err := os.ReadFile(path)
	if err == nil {
		existing = string(data)
	}

	// Idempotency check
	if strings.Contains(existing, "See .pm/AGENTS.md") {
		fmt.Printf("✓ %s already configured\n", path)
		return
	}

	// Create parent directories if needed
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString(pointerBlock); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", path, err)
		os.Exit(1)
	}

	if existing == "" {
		fmt.Printf("✓ Created %s\n", path)
	} else {
		fmt.Printf("✓ Updated %s\n", path)
	}
}

func init() {
	aiInitCmd.Flags().String("for", "", "Tool to configure: claude, copilot, cursor, aider, all (default: all)")
	aiCmd.AddCommand(aiInitCmd)
}
