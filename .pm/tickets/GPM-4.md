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
# Result: labels: [bug, critical]  (array, REPLACES existing labels)

pm edit GPM-1 --field depends_on=GPM-2,GPM-3
# Result: depends_on: [GPM-2, GPM-3]  (array)

pm edit GPM-1 --field points=5
# Result: points: 5  (integer, not string)
```

## Behavior Specification

### Replacement, Not Append

**Important:** `--field` performs REPLACEMENT, not append/merge:

```bash
# Given a ticket with: labels: [auth, backend]
pm edit GPM-1 --field labels=frontend,ui

# Result: labels: [frontend, ui]  (auth and backend are REMOVED)
```

**Rationale:**
- Consistent with single-value field behavior (e.g., `--field assignee=bob` replaces, doesn't add)
- Simple and predictable
- Users can read current values first if they want to preserve them

**For append behavior:**
- User must read ticket first: `pm show GPM-1`
- Then explicitly include all values: `--field labels=auth,backend,frontend,ui`
- Future enhancement: add `--append` flag (e.g., `--field labels=frontend --append`)

### Delimiter Handling

**Only comma (`,`) is supported as delimiter:**

```bash
# Valid:
pm edit GPM-1 --field labels=bug,critical,p1
pm edit GPM-1 --field depends_on=GPM-2,GPM-3,GPM-4

# Invalid (treated as single value):
pm edit GPM-1 --field labels=bug;critical
# Result: labels: ["bug;critical"]  (single element with semicolon in it)

pm edit GPM-1 --field labels=bug|critical
# Result: labels: ["bug|critical"]  (single element with pipe in it)
```

**Edge cases:**
- **Comma in value**: Not supported (YAML doesn't need escaping, but our parser splits on comma)
  ```bash
  pm edit GPM-1 --field labels=bug,needs-review
  # Result: labels: [bug, needs-review]  (correct)
  
  # If you need a label with comma (rare), edit the file directly
  ```
- **Trailing/leading whitespace**: Automatically trimmed
  ```bash
  pm edit GPM-1 --field labels=" bug , critical "
  # Result: labels: [bug, critical]  (spaces trimmed)
  ```
- **Empty elements**: Ignored
  ```bash
  pm edit GPM-1 --field labels=bug,,critical
  # Result: labels: [bug, critical]  (empty string filtered out)
  ```
- **Single element**: Still becomes array
  ```bash
  pm edit GPM-1 --field labels=bug
  # Result: labels: [bug]  (array with one element)
  ```
- **Empty value**: Clears the array
  ```bash
  pm edit GPM-1 --field labels=
  # Result: labels: []  (empty array)
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
  - [ ] Define array fields: `labels`, `depends_on`, `blocks`, `related`
  - [ ] Define integer fields: `points`
  - [ ] Define string fields: `title`, `assignee`, `parent`
  - [ ] Define enum fields with validation rules
- [ ] Update `cmd/pm/edit.go::updateTicketField()` to parse by type
  - [ ] Array parsing: `strings.Split()` on comma, trim whitespace, filter empty
  - [ ] Integer parsing: `strconv.Atoi()` with error handling
  - [ ] Enum validation: check against allowed values
- [ ] Add validation for enum fields (status must be in workflow.yaml)
- [ ] Marshal values to YAML with correct types
- [ ] Add tests for each field type
  - [ ] Test array replacement behavior
  - [ ] Test delimiter handling (comma only)
  - [ ] Test whitespace trimming
  - [ ] Test empty value clearing
  - [ ] Test invalid delimiters (become single value)
- [ ] Update help text with examples and delimiter note

## Acceptance Criteria

- [x] `--field labels=a,b,c` creates array `[a, b, c]` and REPLACES existing labels
- [ ] `--field labels=a` creates array `[a]` (single element)
- [ ] `--field labels=` creates empty array `[]`
- [ ] Comma is the only delimiter; semicolons/pipes treated as part of value
- [ ] Leading/trailing whitespace is trimmed from array elements
- [ ] Empty elements (from `,,`) are filtered out
- [ ] `--field points=5` creates integer `5`, not string "5"
- [ ] `--field status=invalid` rejects with error message
- [ ] Help text documents replacement behavior and comma delimiter
- [ ] All existing functionality still works

