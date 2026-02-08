---
id: GPM-36
title: "Add unit tests for cmd/pm/edit.go"
type: task
status: done
priority: high
points: 3

parent: "GPM-30"
depends_on: []
blocks: []
related: []

labels: [testing, unit-tests]
assignee: ""
created_at: "2026-02-08T08:16:34Z"
updated_at: "2026-02-08T08:16:34Z"
---

# Description

[Claude Opus 4.6]

Add unit tests for edit.go which has critical field-parsing logic used by multiple commands (assign, move).

## Functions to Test

- `updateTicketField()` - Core field update logic
- YAML marshaling/unmarshaling
- Field type parsing (strings, integers, arrays, enums)
- Timestamp update on modifications

## Create cmd/pm/edit_test.go with tests for:

- [x] Update string fields (assignee, title)
- [x] Update integer fields (points)
- [x] Update enum fields (priority, status)
- [x] Update array fields (labels, depends_on) - replacements, not appends
- [x] updated_at timestamp always changes
- [x] YAML marshaling preserves format
- [x] Invalid field types handled gracefully

## Acceptance Criteria

- [x] edit_test.go created with comprehensive unit tests
- [x] All field types covered
- [x] All tests pass: `make test`
- [x] No file I/O in unit tests (use temp files where needed)

## Implementation Notes

Created comprehensive unit tests in cmd/pm/edit_test.go with 6 test functions:
- TestUpdateTicketFieldStringField: Tests updating string fields (assignee, title) including empty values and email addresses
- TestUpdateTicketFieldIntegerField: Tests integer field updates (points), including zero values and invalid inputs
- TestUpdateTicketFieldArrayField: Tests array field updates (labels, depends_on, etc.) with focus on replacement semantics (not append)
- TestUpdateTicketFieldEnumField: Tests enum field updates (priority, type, status) with validation against allowed values
- TestUpdateTicketFieldUpdatesTimestamp: Tests that updated_at timestamp changes while created_at remains unchanged
- TestUpdateTicketFieldPreservesYAMLFormat: Tests that YAML structure and markdown body are preserved after updates

All tests passing. Tests use t.TempDir() for isolated file I/O. YAML marshaling/unmarshaling tested with real ticket files.