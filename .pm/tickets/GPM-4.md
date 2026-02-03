---
id: GPM-4
title: "Fix pm edit --field to handle array fields properly"
type: bug
status: backlog
priority: medium
points: 3
parent: ""
depends_on: []
blocks: []
related: []
labels: [bug, usability]
assignee: ""
created_at: "2026-02-03T03:26:08Z"
updated_at: "2026-02-03T03:29:38Z"
---

# Description

[Sonnet 4.5]

**Bug:** `pm edit --field` treats all field values as strings, which breaks array fields like `labels`, `depends_on`, `blocks`, and `related`.

## Current Behavior

```bash
pm edit GPM-1 --field labels=bug,critical
# Result: labels: "bug,usability"  (string, not array)
```

## Expected Behavior

```bash
pm edit GPM-1 --field labels=bug,critical
# Result: labels: [bug, critical]  (array)
```

## Affected Fields

- **Arrays:** `labels`, `depends_on`, `blocks`, `related`
- **Integers:** `points`
- **Enums:** `status` (should validate against workflow), `type`, `priority`

## Solution

Implement field-specific type handling in `updateTicketField()`:

1. Define field types (string, int, array, enum)
2. Parse values based on field type:
   - Arrays: split on comma, trim whitespace
   - Integers: parse as int
   - Enums: validate against allowed values
3. Update YAML with correctly typed values

## Implementation Steps

- [ ] Create field type registry in `internal/ticket/fields.go`
- [ ] Update `cmd/pm/edit.go::updateTicketField()` to parse by type
- [ ] Add validation for enum fields (status must be in workflow.yaml)
- [ ] Add tests for each field type
- [ ] Update documentation with examples

## Acceptance Criteria

- [ ] `--field labels=a,b,c` creates array `[a, b, c]`
- [ ] `--field points=5` creates integer `5`
- [ ] `--field status=invalid` rejects with error
- [ ] All existing functionality still works

