---
id: GPM-36
title: "Add unit tests for cmd/pm/edit.go"
type: task
status: backlog
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

- [ ] Update string fields (assignee, title)
- [ ] Update integer fields (points)
- [ ] Update enum fields (priority, status)
- [ ] Update array fields (labels, depends_on) - replacements, not appends
- [ ] updated_at timestamp always changes
- [ ] YAML marshaling preserves format
- [ ] Invalid field types handled gracefully

## Acceptance Criteria

- [ ] edit_test.go created with comprehensive unit tests
- [ ] All field types covered
- [ ] All tests pass: `make test`
- [ ] No file I/O in unit tests (use temp files where needed)