---
id: GPM-51
title: "Add --parent flag to pm new command"
type: task
status: backlog
priority: medium
points: 1

parent: ""
depends_on: []
blocks: []
related: []

labels: [ux, cli]
assignee: ""
created_at: "2026-02-08T14:58:00Z"
updated_at: "2026-02-08T14:58:00Z"
---

# Description

[Claude Sonnet 4.5]

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
pm new --type task --parent GPM-44 --priority high "Fix cache bug"
✓ Created new ticket: GPM-52 (parent: GPM-44, type: task, priority: high)

# Without parent (existing behavior)
pm new "Standalone ticket"
✓ Created new ticket: GPM-53
```

### Implementation Steps

- [ ] Add `--parent` flag to `pm new` command in `cmd/pm/new.go`
- [ ] Validate parent ticket exists before creating
- [ ] Set `parent` field in generated YAML
- [ ] Update output message to show parent if specified
- [ ] Add shell completion for parent ticket IDs
- [ ] Update help text with examples

### Validation

- Parent ticket must exist (error if not found)
- Parent ticket can be any type (epic, story, task, bug)
- Optional: Warn if creating child of a completed ticket (suggest alternative parent)

## Testing

- [ ] Can create ticket with parent
- [ ] Error if parent doesn't exist
- [ ] Parent field correctly set in YAML
- [ ] Works with other flags (--type, --priority, etc.)
- [ ] Works without parent (unchanged behavior)
- [ ] Shell completion shows available parent tickets

## Acceptance Criteria

- [ ] `pm new --parent <id> <title>` creates child ticket correctly
- [ ] Parent field is set in the YAML front-matter
- [ ] Validation prevents invalid parent references
- [ ] Output message shows parent if specified
- [ ] Works in combination with other flags
- [ ] Shell completion available for parent ticket IDs
- [ ] All tests pass
- [ ] Help text updated with examples
