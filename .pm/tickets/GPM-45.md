---
id: GPM-45
title: "Implement pm link and pm unlink with automatic symmetry"
type: task
status: done
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
updated_at: "2026-02-08T16:45:00Z"
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
- **Idempotent**: Arrays are de-duplicated on write (safe to link multiple times)
- **No auto-commit**: Modifies files, let user commit manually
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

### File I/O Pattern

Follow the existing pattern from `pm edit` and `pm assign`:
1. Read file: `content, err := os.ReadFile(ticketPath)`
2. Parse: Split on `---`, unmarshal YAML front-matter with `yaml.v3`
3. Modify: Update array fields
4. Normalize: Call `ticket.Normalize()` to de-duplicate arrays
5. Serialize: Marshal YAML, update `updated_at`, reconstruct file
6. Write: `os.WriteFile(ticketPath, newContent, 0644)`

- [x] Create `cmd/pm/link.go` with link command
- [x] Create `cmd/pm/unlink.go` with unlink command
- [x] Add `Normalize()` method to ticket struct in internal/ticket/ticket.go:
  - [x] De-duplicate all array fields (depends_on, blocks, related, labels)
  - [x] Call during serialization before file write
- [x] Create `internal/ticket/relationships.go` with helper function:
  - [x] `func updateRelationshipWithSymmetry(sourceID, targetID, relType string, add bool) (bool, error)`
  - [x] Handles load, modify, save for both tickets
  - [x] Implements symmetry logic internally (depends-on ↔ blocks)
  - [x] For `related`, only updates source ticket (unidirectional)
- [x] In link command:
  - [x] Check if relationship already exists (Option A: return "Already linked" message)
  - [x] Validate both ticket IDs exist and aren't self-reference before any file operations
  - [x] Call `updateRelationshipWithSymmetry(..., true)` to add
- [x] In unlink command:
  - [x] Validate ticket IDs exist before any file operations
  - [x] Call `updateRelationshipWithSymmetry(..., false)` to remove
- [x] Add validation layer (before any file modifications):
  - [x] Both ticket IDs must exist
  - [x] No self-reference (GPM-5 cannot depend on GPM-5)
  - [x] Type must be valid (depends-on, blocks, related)
  - [x] Exit immediately on validation failure
- [x] Print informative output with consistent format
- [x] Add shell completion:
  - [x] Ticket IDs: Use existing `completeTicketIDs` function
  - [x] Type flag: Return static list `["depends-on", "blocks", "related"]`

## Examples

```bash
# Create dependency (modifies both tickets)
pm link GPM-5 GPM-10 --type depends-on
✓ Linked GPM-5 -> GPM-10 (depends-on) [modified 2 files]
# GPM-5.md: depends_on: [GPM-10]
# GPM-10.md: blocks: [GPM-5]

# Create blocking relationship (same as above, inverse direction)
# no need to list the exact filenames, it is obvious by the ticket IDs
pm link GPM-10 GPM-5 --type blocks
✓ Linked GPM-10 -> GPM-5 (blocks) [modified 2 files]

# Create loose association (only modifies source)
pm link GPM-5 GPM-99 --type related
✓ Linked GPM-5 -> GPM-99 (related) [modified 1 file]
# GPM-5.md: related: [GPM-99]
# GPM-99.md: unchanged

# Link twice (idempotent - first time adds, second time already exists)
pm link GPM-5 GPM-10 --type depends-on
✓ Linked GPM-5 -> GPM-10 (depends-on) [modified 2 files]
pm link GPM-5 GPM-10 --type depends-on
✓ Already linked: GPM-5 -> GPM-10 (depends-on)

# Remove dependency (removes inverse too)
pm unlink GPM-5 GPM-10 --type depends-on
✓ Unlinked GPM-5 -> GPM-10 (depends-on) [modified 2 files]

# Remove from all relationships (source only)
pm unlink GPM-5 GPM-99
✓ Removed GPM-99 from all relationships in GPM-5 [modified 1 file]

# Error handling
pm link GPM-5 GPM-5 --type depends-on
Error: Cannot link ticket to itself (GPM-5)

pm link GPM-5 NONEXISTENT --type depends-on
Error: Target ticket not found: NONEXISTENT
```

## Edge Cases

- **Target ticket doesn't exist**: Error with helpful message
- **Relationship already exists**: No-op, print message `✓ Already linked: GPM-5 -> GPM-10 (depends-on)`
- **Unlink non-existent relationship**: No-op, print message `✓ Not linked: GPM-5 -> GPM-10 (depends-on)`
- **Self-reference attempt**: Error `Error: Cannot link ticket to itself (GPM-5)`
- **Invalid type**: Error `Error: Invalid type 'foo' - must be depends-on, blocks, or related`
- **File parse errors**: Gracefully handle and report YAML errors
- **Array uniqueness**: Arrays are automatically de-duplicated on write (silent normalization)
  - If user manually adds duplicates in YAML, next command that saves the file will fix it
  - Commands become naturally idempotent without explicit duplicate checks

## Error Handling

### Validation-First Pattern

- **Validate before any file modifications**: Check both tickets exist, no self-reference, valid type
- **Exit immediately on error**: Stop execution, don't proceed to modify files
- **Clear error messages**: Include ticket ID and helpful context

### Atomic Symmetry Operations

- **All-or-nothing**: If modifying both source and target (symmetry), either both succeed or neither
- **Rollback on failure**:
  - Update source ticket ✓
  - Update target ticket ✗ (file permission, parse error, etc.)
  - **Action**: Revert source ticket to original state
  - **Report**: `Error: Failed to link GPM-5 -> GPM-10 (couldn't write GPM-10). Reverted changes to GPM-5.`
- **Unidirectional operations** (e.g., `related`) can fail safely: only one file involved

### Output Format (Consistent)

- **Success**: `✓ Linked GPM-5 -> GPM-10 (depends-on) [modified 2 files]`
- **Already linked**: `✓ Already linked: GPM-5 -> GPM-10 (depends-on)`
- **Not linked**: `✓ Not linked: GPM-5 -> GPM-10 (depends-on)`
- **Error**: `Error: <specific reason>`

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

