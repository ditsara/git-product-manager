---
id: GPM-16
title: "Improve help messages for commands with missing arguments"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 2  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""  # Parent epic or story
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: [ux, cli]  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-03T05:34:53Z"
updated_at: "2026-02-03T05:34:53Z"
---

# Description

[Sonnet 4.5]

Improve the user experience when commands are run without required arguments or with no arguments at all.

## Current Problem

- Running `pm` with no arguments shows generic Cobra usage
- Running `pm show` with no ticket ID shows an error but could be more helpful
- New users don't know what commands are available without reading docs

## Solution

**When `pm` is run with no arguments:**
- Display available commands (similar to `git` or `docker`)
- Show brief description of each command
- Suggest `pm --help` for more details

**When commands are missing required arguments:**
- Show usage example for that specific command
- Include the expected argument format
- Provide a common example (e.g., `pm show GPM-123`)

## Implementation Steps

- [ ] Update `cmd/pm/main.go` to customize root command's Run function
- [ ] Add helpful error messages in `show.go`, `move.go`, `edit.go` for missing args
- [ ] Use Cobra's `Args` validator functions (e.g., `cobra.ExactArgs(1)`)
- [ ] Add examples to each command's `Long` description

## Acceptance Criteria

- [ ] Running `pm` with no args shows list of available commands
- [ ] Running `pm show` with no ID shows usage example
- [ ] Running `pm move` with missing args shows usage example
- [ ] All error messages are helpful and actionable

## Examples

**Before:**
```
$ pm
Error: required flag(s) "..." not set
```

**After:**
```
$ pm
Usage: pm [command]

Available Commands:
  init        Initialize .pm directory in a project
  new         Create a new ticket
  list        List tickets
  show        Display ticket details
  edit        Edit a ticket
  move        Change ticket status

Use "pm [command] --help" for more information about a command.
```

**Before:**
```
$ pm show
Error: accepts 1 arg(s), received 0
```

**After:**
```
$ pm show
Error: ticket ID required

Usage: pm show <ticket-id>

Example:
  pm show GPM-123
```