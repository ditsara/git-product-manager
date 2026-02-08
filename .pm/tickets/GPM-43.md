---
id: GPM-43
title: "UX Polish & CLI Refinements"
type: epic
status: done
priority: high
points: 0

parent: ""
depends_on: []
blocks: []
related: []

labels: [ux, cli, polish, enhancement]
assignee: ""
created_at: "2026-02-08T09:40:28Z"
updated_at: "2026-02-08T11:35:00Z"
---

# Description

[Claude Sonnet 4.5]

Improve the day-to-day user experience of the `pm` CLI through polish, better feedback, and discoverability enhancements.

## Scope

This epic groups user-facing improvements that make the tool more pleasant and intuitive to use:

- **Shell integration**: Tab completion for commands, ticket IDs, and flags
- **Better error handling**: Clear messages with proper formatting
- **Improved help**: Contextual guidance when arguments are missing
- **Visual indicators**: Show ticket hierarchy at a glance
- **Smart filtering**: Hide completed work by default, reduce list clutter

## Goals

1. **Reduce friction**: Common tasks should feel natural and fast
2. **Discoverability**: Users should discover features through usage, not documentation
3. **Polish**: Small details matter (trailing newlines, column alignment, etc.)
4. **Consistency**: Error messages, help text, and output should follow patterns

## Child Tickets

- GPM-42: Bash completion for ticket IDs and commands
- GPM-40: Error messages missing trailing newline
- GPM-16: Improve help messages for missing arguments
- GPM-26: Visual indicator for tickets with children
- GPM-39: State groups & filter completed tickets by default

## Success Criteria

- Tab completion works for all major shells (bash, zsh, fish)
- Error output is clean and properly formatted
- New users can discover commands without reading docs
- `pm list` is focused and actionable (hides completed work)
- Ticket hierarchy is visible without extra commands

## Stage

This represents "Stage 1.7" - UX polish after core functionality is stable.

## Implementation Summary

All child tickets have been completed:

✅ **GPM-42**: Bash completion implemented for all shells (bash, zsh, fish, powershell)
- Ticket ID completion for all relevant commands
- State completion for `pm move` and `--status` flag
- Type completion for `--type` flag
- Case-insensitive matching
- Graceful error handling

✅ **GPM-40**: Error messages now end with trailing newline
- Fixed prompt formatting issue in main.go
- Tested with various error conditions

✅ **GPM-16**: Improved help messages for missing arguments
- Commands show helpful usage when arguments are missing
- Better user guidance

✅ **GPM-26**: Visual indicator for tickets with children
- Shows `(+)` after ticket ID when ticket has children
- Clean hierarchy visualization without extra columns

✅ **GPM-39**: State groups & smart filtering
- `pm list` now excludes completed tickets by default
- New flags: `--completed`, `--active`, `--incomplete`
- Configurable via `state_groups` in workflow.yaml
- Backward compatible

**Result**: The CLI is now polished, discoverable, and user-friendly. All 239 tests passing.
