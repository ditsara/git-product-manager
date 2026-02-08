---
id: GPM-42
title: "Implement bash completion for ticket IDs and commands"
type: task
status: done
priority: high
points: 3

parent: GPM-43
depends_on: []
blocks: []
related: []

labels: [ux, cli, completion]
assignee: ""
created_at: "2026-02-08T09:37:05Z"
updated_at: "2026-02-08T11:30:00Z"
---

# Description

[Claude Sonnet 4.5]

Implement shell completion support for better CLI UX, with intelligent completion of ticket IDs, commands, flags, and flag values.

## User Request

Enable tab completion for ticket IDs in commands. For example:

```bash
pm list --parent gpm-1<TAB>
# Should complete to: GPM-1, GPM-10, GPM-11, GPM-12, etc.

pm show gpm-<TAB>
# Should show all available ticket IDs starting with GPM-

pm move GPM-1 <TAB>
# Should show valid states from workflow.yaml: backlog, todo, in-progress, done, canceled
```

## Cobra's Built-in Completion Support

Cobra provides built-in completion generation for bash, zsh, fish, and PowerShell. We need to:

1. Add `completion` command to generate completion scripts
2. Register custom completion functions for ticket IDs
3. Register custom completion functions for workflow states
4. Register custom completion functions for other enums (types, labels, etc.)

## Implementation Steps

### 1. Add Completion Command

- [x] Create `cmd/pm/completion.go` with completion subcommand
- [x] Use Cobra's `GenBashCompletion()`, `GenZshCompletion()`, etc.
- [x] Support all shells: bash, zsh, fish, powershell
- [x] Add usage instructions to help text

### 2. Custom Completion Functions

Cobra supports `ValidArgsFunction` for dynamic completions:

```go
// Example: Complete ticket IDs
func completeTicketIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    ticketsPath := filepath.Join(".pm", "tickets")
    files, err := os.ReadDir(ticketsPath)
    if err != nil {
        return nil, cobra.ShellCompDirectiveNoFileComp
    }
    
    var tickets []string
    prefix := strings.ToUpper(toComplete) // Case-insensitive
    
    for _, file := range files {
        if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
            continue
        }
        ticketID := strings.TrimSuffix(file.Name(), ".md")
        
        // Filter by prefix if provided
        if toComplete == "" || strings.HasPrefix(strings.ToUpper(ticketID), prefix) {
            tickets = append(tickets, ticketID)
        }
    }
    
    return tickets, cobra.ShellCompDirectiveNoFileComp
}
```

### 3. Register Completions for Each Command

- [x] `pm show <TAB>` → complete ticket IDs
- [x] `pm edit <TAB>` → complete ticket IDs
- [x] `pm move <id> <TAB>` → complete workflow states
- [x] `pm assign <TAB>` → complete ticket IDs
- [x] `pm comment <TAB>` → complete ticket IDs
- [x] `pm history <TAB>` → complete ticket IDs
- [x] `pm list --parent <TAB>` → complete ticket IDs
- [x] `pm list --status <TAB>` → complete workflow states
- [x] `pm new --type <TAB>` → complete types (story, task, bug, epic)
- [x] `pm list --label <TAB>` → complete labels from labels.yaml

### 4. State Completion Function

```go
func completeStates(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    workflow, err := config.LoadWorkflow(".pm")
    if err != nil {
        return nil, cobra.ShellCompDirectiveNoFileComp
    }
    
    var states []string
    for _, state := range workflow.States {
        if strings.HasPrefix(state, toComplete) {
            states = append(states, state)
        }
    }
    
    return states, cobra.ShellCompDirectiveNoFileComp
}
```

### 5. Update Each Command

Add `ValidArgsFunction` to commands:

```go
var showCmd = &cobra.Command{
    Use:   "show <id>",
    Short: "Show a single ticket",
    Args:  cobra.ExactArgs(1),
    ValidArgsFunction: completeTicketIDs,
    Run: func(cmd *cobra.Command, args []string) {
        // existing code
    },
}
```

For flags:

```go
listCmd.Flags().String("status", "", "Filter by status")
listCmd.RegisterFlagCompletionFunc("status", completeStates)

listCmd.Flags().String("parent", "", "Filter by parent ticket")
listCmd.RegisterFlagCompletionFunc("parent", completeTicketIDs)
```

### 6. Completion Command Implementation

```go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate shell completion script",
    Long: `Generate shell completion script for pm.

To load completions:

Bash:
  $ source <(pm completion bash)
  
  # To load completions for each session, add to ~/.bashrc:
  $ pm completion bash > ~/.pm-completion.sh
  $ echo 'source ~/.pm-completion.sh' >> ~/.bashrc

Zsh:
  $ source <(pm completion zsh)
  
  # To load completions for each session, add to ~/.zshrc:
  $ pm completion zsh > "${fpath[1]}/_pm"

Fish:
  $ pm completion fish | source
  
  # To load completions for each session:
  $ pm completion fish > ~/.config/fish/completions/pm.fish

PowerShell:
  PS> pm completion powershell | Out-String | Invoke-Expression
  
  # To load completions for each session, add to your PowerShell profile:
  PS> pm completion powershell >> $PROFILE
`,
    DisableFlagsInUseLine: true,
    ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
    Args:                  cobra.ExactValidArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        switch args[0] {
        case "bash":
            cmd.Root().GenBashCompletion(os.Stdout)
        case "zsh":
            cmd.Root().GenZshCompletion(os.Stdout)
        case "fish":
            cmd.Root().GenFishCompletion(os.Stdout, true)
        case "powershell":
            cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
        }
    },
}
```

## Testing

Manual testing required (completion is shell-specific):

```bash
# Build and install
go build -o bin/pm ./cmd/pm

# Test bash completion
source <(./bin/pm completion bash)

# Test completions
pm show gpm-<TAB>      # Should list GPM-1, GPM-10, etc.
pm list --parent <TAB>  # Should list all ticket IDs
pm move GPM-1 <TAB>     # Should list: backlog, todo, in-progress, done, canceled
pm new --type <TAB>     # Should list: story, task, bug, epic
```

## Edge Cases

- **Not in a pm project**: Completion should gracefully handle missing `.pm/` directory
- **Empty tickets directory**: Return no completions (not an error)
- **Case sensitivity**: Ticket ID completion should be case-insensitive (`gpm-1` matches `GPM-1`)
- **Partial matches**: `pm show gpm-1<TAB>` should show `GPM-1`, `GPM-10`, `GPM-11`, etc.
- **Invalid workflow.yaml**: State completion should fail gracefully

## Documentation Updates

- [x] Update README.md with completion installation instructions
- [x] Add examples to help text
- [x] Document in AGENTS.md (move from "Future Improvements" to implemented)

## Acceptance Criteria

- [x] `pm completion bash|zsh|fish|powershell` generates working completion scripts
- [x] Ticket ID completion works for all relevant commands
- [x] State completion works for `pm move` and `--status` flag
- [x] Type completion works for `--type` flag
- [x] Completion is case-insensitive for ticket IDs
- [x] Gracefully handles missing `.pm/` directory
- [x] Help text includes installation instructions for each shell
- [x] README.md updated with completion setup
- [x] All manual tests pass

## References

- Cobra completion docs: https://github.com/spf13/cobra/blob/main/shell_completions.md
- AGENTS.md section on "Bash Completion (Future)"
- README.md already has placeholder for shell completion

