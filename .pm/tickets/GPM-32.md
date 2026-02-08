---
id: GPM-32
title: "Split assign tests into integration_assign_test.go"
type: task
status: done
priority: medium
points: 1

parent: "GPM-30"
depends_on: []
blocks: []
related: []

labels: [testing, refactoring]
assignee: ""
created_at: "2026-02-08T08:16:35Z"
updated_at: "2026-02-08T08:16:35Z"
---

# Description

[Claude Opus 4.6]

Extract all assign command tests from integration_test.go into a dedicated test file.

## Tests to Move

- TestAssignTicket
- TestAssignTicketIdempotent
- TestAssignTicketWithEmail
- TestAssignTicketCaseInsensitive
- TestAssignTicketUpdateTimestamp

## Implementation Steps

- [x] Create `integration_assign_test.go`
- [x] Move all assign-related tests from integration_test.go
- [x] Ensure helper functions (initWorkspace, runPM) are accessible
- [x] Remove moved tests from integration_test.go
- [x] Verify all tests pass: `make test`

## Acceptance Criteria

- [x] New file has ~150 lines with 5 assign tests (actually 157 lines)
- [x] All tests pass
- [x] No broken references

## Implementation Notes

- All 5 assign tests successfully moved to `integration_assign_test.go`
- Removed unused `ticket` import from integration_test.go
- All tests passing