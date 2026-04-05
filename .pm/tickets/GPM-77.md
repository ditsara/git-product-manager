---
assignee: ""
blocks: []
created_at: "2026-04-05T08:13:57Z"
depends_on: []
id: GPM-77
labels: []
parent: ""
points: 0
priority: medium
related: []
status: done
title: Simplify shell integration
type: story
updated_at: "2026-04-05T08:23:33Z"
---


# Description

Currently, the completion help reads (bash example):

```
Bash:
  $ source <(pm completion bash)
  
  # To load completions for each session, add to ~/.bashrc:
  $ pm completion bash > ~/.pm-completion.sh
  $ echo 'source ~/.pm-completion.sh' >> ~/.bashrc
```

If you do `pm completion`, it only gives you an error message.

Can we make it so `pm completion` displays the full help message automatically?

Also, simplify the shell integration instructions to something like:

put this in your `~/.bashrc`:

```
eval $(pm completion bash)
```

So the user doesn't need to generate an intermediate file that can go out of
date. Similar to how rbenv shell integration works.

# Implementation Notes

**File:** `cmd/pm/completion.go`

Two changes, both in `completionCmd`:

1. **`pm completion` (no args) shows help instead of erroring**
   - Change `Args: cobra.ExactValidArgs(1)` → `Args: cobra.MaximumNArgs(1)`
   - In the `Run` func, add an early return that calls `cmd.Help()` when `len(args) == 0`
   - Add a `default` case to the switch that prints an error and exits non-zero for unknown shell names (since `ExactValidArgs` previously handled that validation)

2. **Simplify the `Long` help text** to `eval "$(pm completion <shell>)"` style per shell:
   - Bash (`~/.bashrc`): `eval "$(pm completion bash)"`
   - Zsh (`~/.zshrc`): `eval "$(pm completion zsh)"`
   - Fish (`~/.config/fish/config.fish`): `pm completion fish | source`
   - PowerShell (`$PROFILE`): `pm completion powershell | Out-String | Invoke-Expression`

# Acceptance Criteria

- `pm completion` (no args) prints the help/usage text and exits 0
- `pm completion bash` still emits a valid bash completion script
- `pm completion invalid` exits non-zero with a clear error message
- Help text shows the simplified `eval "$(pm completion <shell>)"` one-liner per shell
