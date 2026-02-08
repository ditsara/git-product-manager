---
id: GPM-51
title: "Add --parent flag to pm new command"
type: task
status: done
priority: medium
points: 1

parent: ""
depends_on: []
blocks: []
related: []

labels: [ux, cli]
assignee: ""
created_at: "2026-02-08T14:58:00Z"
updated_at: "2026-02-08T15:50:00Z"
---

# Description

[Claude Haiku 4.5]

Add `--parent` flag to `pm new` command to allow specifying a parent ticket when creating a new ticket, improving workflow efficiency.

## Problem Statement

Creating a child ticket currently requires two operations:

```bash
pm new "Implement feature X"      # Create ticket, get ID (e.g., GPM-51)
pm edit GPM-51 --field parent=GPM-2  # Set parent
```

This is tedious for common workflows like breaking down epics into stories.

## Solution

Add optional `--parent` flag to `pm new`:

```bash
pm new --parent GPM-2 "Implement feature X"
# ✓ Created new ticket: GPM-51 (parent: GPM-2)
```

## Implementation

### Command Usage

```bash
# Basic usage with parent
pm new --parent GPM-44 "Validate YAML input"
✓ Created new ticket: GPM-51 (parent: GPM-44)

# With other options
pm new --type task --parent GPM-44 "Fix cache bug"
✓ Created new ticket: GPM-52 (parent: GPM-44, type: task)

# Without parent (existing behavior)
pm new "Standalone ticket"
✓ Created new ticket: GPM-53
```

### Implementation Steps

- [x] Add `--parent` flag to `pm new` command in `cmd/pm/new.go`
- [x] Validate parent ticket exists before creating
- [x] Set `parent` field in generated YAML
- [x] Update output message to show parent if specified
- [x] Add shell completion for parent ticket IDs
- [x] Update help text with examples
- [x] Update all templates (story, task, bug, epic) to use {{.Parent}}

### Validation

- Parent ticket must exist (error if not found)
- Parent ticket can be any type (epic, story, task, bug)

## Edge Cases
- Parent doesn't exist - error with helpful message ✓
- User self-references parent - prevented by ticket ID not existing before creation ✓
- No parent specified - empty string renders correctly ✓

## Testing

- [x] Can create ticket with parent
- [x] Error if parent doesn't exist
- [x] Parent field correctly set in YAML
- [x] Works with other flags (--type)
- [x] Works without parent (unchanged behavior)
- [x] Shell completion shows available parent tickets

## Acceptance Criteria

- [x] `pm new --parent <id> <title>` creates child ticket correctly
- [x] Parent field is set in the YAML front-matter
- [x] Validation prevents invalid parent references
- [x] Output message shows parent if specified
- [x] Works in combination with other flags
- [x] Shell completion available for parent ticket IDs
- [x] All tests pass (15/15 unit tests passing)
- [x] Help text updated with examples

## Implementation Notes

**Key Implementation Details:**
- Added `ticketExists()` helper function to validate parent ticket existence
- Updated all four template files (both in `cmd/pm/templates/` and `.pm/config/templates/`)
- Templates use `{{.Parent}}` template variable for parent field
- Flag defined as `-p/--parent` for consistency with other CLI tools
- Validation happens BEFORE ticket creation (no orphaned tickets)
- Output message conditionally includes parent info when specified
- Shell completion reuses existing `completeTicketIDs()` function
