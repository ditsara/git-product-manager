---
id: GPM-45
title: "Implement pm link and pm unlink with automatic symmetry"
type: task
status: backlog  # Current workflow state
priority: high  # low, medium, high, critical
points: 5  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: GPM-2  # Parent epic or story
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: [GPM-46]  # Related work (duplicates, see-also)

labels: [relationships, cli]  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-08T14:38:49Z"
updated_at: "2026-02-08T14:38:49Z"
---

# Description

[Claude Sonnet 4.5]

Implement `pm link` and `pm unlink` commands to manage array-based ticket relationships with automatic bidirectional symmetry for `depends-on`/`blocks` pairs.

## Command Specifications

### `pm link <id> <target-id> [--type TYPE]`

**Supported Types:**
- `depends-on`: Creates dependency relationship (bidirectional with blocks)
- `blocks`: Creates blocking relationship (bidirectional with depends-on)
- `related`: Creates loose association (unidirectional)
- **Default**: `related`

**Behavior:**
- Appends target ID to the specified array field in source ticket
- **Automatic Symmetry**: If type is `depends-on` or `blocks`, also updates the inverse relationship:
  - `link A B --type depends-on` → Adds B to A's depends_on AND adds A to B's blocks
  - `link A B --type blocks` → Adds B to A's blocks AND adds A to B's depends_on
- **Idempotent**: Won't duplicate if relationship already exists
- **No auto-commit**: Modifies files, user commits when ready
- **Output**: `✓ Linked GPM-5 -> GPM-10 (depends-on) [modified 2 files]`

### `pm unlink <id> <target-id> [--type TYPE]`

**Behavior:**
- Removes target ID from the specified array field in source ticket
- **Automatic Symmetry**: If type is `depends-on` or `blocks`, also removes inverse:
  - `unlink A B --type depends-on` → Removes B from A's depends_on AND removes A from B's blocks
  - `unlink A B --type blocks` → Removes B from A's blocks AND removes A from B's depends_on
- **No type specified**: Removes from all array fields (depends_on, blocks, related) in source ticket only
- **Output**: `✓ Unlinked GPM-5 -> GPM-10 (depends-on) [modified 2 files]`

## Implementation Steps

- [ ] Create `cmd/pm/link.go` with link command
- [ ] Create `cmd/pm/unlink.go` with unlink command
- [ ] Add helper function `updateRelationship(ticketID, targetID, relType, add bool)` in internal/ticket/
- [ ] Implement automatic symmetry logic:
  - [ ] Map depends-on ↔ blocks as inverse pairs
  - [ ] For related, only update source ticket (unidirectional)
- [ ] Add validation:
  - [ ] Both ticket IDs must exist
  - [ ] No self-reference (GPM-5 cannot depend on GPM-5)
  - [ ] Type must be valid (depends-on, blocks, related)
- [ ] Print informative output showing files modified
- [ ] Add shell completion for ticket IDs and types

## Examples

```bash
# Create dependency (modifies both tickets)
pm link GPM-5 GPM-10 --type depends-on
✓ Linked GPM-5 -> GPM-10 (depends-on) [modified 2 files]
# GPM-5.md: depends_on: [GPM-10]
# GPM-10.md: blocks: [GPM-5]

# Create blocking relationship (same as above, inverse direction)
pm link GPM-10 GPM-5 --type blocks
✓ Linked GPM-10 -> GPM-5 (blocks) [modified 2 files]

# Create loose association (only modifies source)
pm link GPM-5 GPM-99 --type related
✓ Linked GPM-5 -> GPM-99 (related) [modified 1 file]
# GPM-5.md: related: [GPM-99]
# GPM-99.md: unchanged

# Remove dependency (removes inverse too)
pm unlink GPM-5 GPM-10 --type depends-on
✓ Unlinked GPM-5 -> GPM-10 (depends-on) [modified 2 files]

# Remove from all relationships (source only)
pm unlink GPM-5 GPM-99
✓ Removed GPM-99 from all relationships in GPM-5 [modified 1 file]
```

## Edge Cases

- **Target ticket doesn't exist**: Error with helpful message
- **Relationship already exists**: No-op, print message "Already linked"
- **Unlink non-existent relationship**: No-op, print message "Not linked"
- **Self-reference attempt**: Error "Cannot link ticket to itself"
- **Invalid type**: Error "Invalid type: must be depends-on, blocks, or related"
- **File parse errors**: Gracefully handle and report YAML errors

## Testing

- [ ] Unit tests for relationship helper functions
- [ ] Integration test: link depends-on, verify both files updated
- [ ] Integration test: link blocks, verify symmetry
- [ ] Integration test: link related, verify only source updated
- [ ] Integration test: unlink with type, verify both files
- [ ] Integration test: unlink without type, verify all arrays checked
- [ ] Test idempotency (link twice, unlink once)
- [ ] Test validation (self-reference, missing tickets, invalid types)

## Acceptance Criteria

- [ ] `pm link` creates relationships with automatic symmetry for depends-on/blocks
- [ ] `pm unlink` removes relationships with automatic symmetry
- [ ] Commands are idempotent (safe to run multiple times)
- [ ] Clear output shows which files were modified
- [ ] Validation prevents self-reference and missing tickets
- [ ] All tests pass
- [ ] Shell completion works for ticket IDs and types

