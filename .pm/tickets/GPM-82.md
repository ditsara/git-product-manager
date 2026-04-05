---
id: GPM-82
title: "Reorganize pm guide under pm ai subcommand"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-81"  # Parent epic or story
depends_on: []  # Must complete these first
blocks:
  - GPM-83
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-04-05T09:12:26Z"
updated_at: "2026-04-05T09:12:26Z"
---

# Description

Replace the top-level `pm guide` command with `pm ai guide`, grouping all LLM-related commands under a `pm ai` subcommand. `pm guide` is removed entirely — no alias, no deprecation warning (zero adoption, prototype stage).

## What Changes

| Before | After |
|--------|-------|
| `pm guide` | `pm ai guide` |
| `pm guide [section]` | `pm ai guide [section]` |
| `pm guide workflow` | `pm ai guide workflow` |

The `internal/guide` package is unchanged — only the Cobra command wiring moves.

## Implementation

**1. Delete `cmd/pm/guide.go`**

Remove the file entirely.

**2. Create `cmd/pm/ai.go`**

Define a `pm ai` parent command with no `Run` of its own (shows help by default):

```go
var aiCmd = &cobra.Command{
    Use:   "ai",
    Short: "Commands for AI/LLM assistant integration",
}

func init() {
    rootCmd.AddCommand(aiCmd)
}
```

**3. Create `cmd/pm/ai_guide.go`**

Move the guide logic from `guide.go` into a new file as `aiGuideCmd`, registered under `aiCmd`:

```go
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
    ...
}

func init() {
    aiCmd.AddCommand(aiGuideCmd)
}
```

**4. Update `completion_helpers.go`**

`completeGuideSections` currently references `guideCmd`. Update it to attach to `aiGuideCmd`, or make it tool-agnostic.

## Acceptance Criteria

- [ ] `pm guide` no longer exists (returns "unknown command" error)
- [ ] `pm ai guide` works identically to the old `pm guide`
- [ ] `pm ai guide [section]` tab-completes section names
- [ ] `pm ai --help` lists `guide` and (once added) `init` as subcommands
- [ ] `make test` passes
