---
id: GPM-31
title: "Split comment tests into integration_comment_test.go"
type: task
status: done
priority: medium
points: 2

parent: "GPM-30"
depends_on: []
blocks: []
related: []

labels: [testing, refactoring]
assignee: ""
created_at: "2026-02-08T08:16:34Z"
updated_at: "2026-02-08T08:16:34Z"
---

# Description

[Claude Opus 4.6]

Extract all comment-related tests from integration_test.go into a new dedicated test file.

## Tests to Move

- TestCommentDirectMode
- TestCommentWithCustomAuthor
- TestCommentOnNonexistentTicket
- TestMultipleComments
- TestCommentEmptyMessage
- TestCommentFilenameFormat
- TestCommentCaseInsensitive
- TestCommentAmendDirectMode
- TestCommentAmendByTimestamp

## Implementation Steps

- [x] Create `integration_comment_test.go`
- [x] Move all comment-related tests from integration_test.go
- [x] Ensure all helper functions used by tests are copied or kept accessible
- [x] Remove moved tests from integration_test.go
- [x] Verify all tests still pass: `make test`

## Acceptance Criteria

- [x] New file has ~300 lines with 9 comment-related tests (actually 330 lines)
- [x] integration_test.go is reduced by ~300 lines (reduced from 1,585 to 1,251 lines, -334 lines)
- [x] All tests pass
- [x] No broken references or missing helpers

## Note

After all split tickets (GPM-31 through GPM-34 and GPM-38) are completed, only TestIntegrationWorkflow and TestIntegrationInitValidation should remain in integration_test.go (~325 lines including helpers).

## Implementation Notes

- Removed unused `fmt` import from integration_test.go
- All 9 comment tests successfully moved to `integration_comment_test.go`
- Helper functions (buildPMBinary, runPM, initWorkspace) remain in integration_test.go and are accessible to both files