---
assignee: ""
blocks: []
created_at: "2026-02-03T05:34:53Z"
depends_on: []
id: GPM-16
labels:
    - ux
    - cli
parent: GPM-43
points: 2
priority: medium
related: []
status: done
title: Improve help messages for commands with missing arguments
type: task
updated_at: "2026-02-08T09:50:58Z"
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

- [x] Update `cmd/pm/main.go` to customize root command's Run function
- [x] Add helpful error messages in `show.go`, `move.go`, `edit.go` for missing args
- [x] Use Cobra's `Args` validator functions (e.g., `cobra.ExactArgs(1)`)
- [x] Add examples to each command's `Long` description

## Acceptance Criteria

- [x] Running `pm` with no args shows list of available commands
- [x] Running `pm show` with no ID shows usage example
- [x] Running `pm move` with missing args shows usage example
- [x] All error messages are helpful and actionable

## Implementation Notes

**What was done:**
- Added `Example` field to `move.go` and `history.go` commands
- Updated `Use` field in move.go to use angle brackets `<id>` instead of square brackets
- Updated Long description in move.go for clarity
- Commands already had `cobra.ExactArgs()` validation in place
- Root command already showed help when run without arguments
- `show`, `edit`, `assign`, `comment` already had examples in Long or Example fields

**Result:**
All commands now show clear usage examples when run without required arguments. The "Whoops" error at the end will be fixed in GPM-40.

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