---
id: GPM-31
title: "Split comment tests into integration_comment_test.go"
type: task
status: backlog
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

- [ ] Create `integration_comment_test.go`
- [ ] Move all comment-related tests from integration_test.go
- [ ] Ensure all helper functions used by tests are copied or kept accessible
- [ ] Remove moved tests from integration_test.go
- [ ] Verify all tests still pass: `make test`

## Acceptance Criteria

- [ ] New file has ~300 lines with 9 comment-related tests
- [ ] integration_test.go is reduced by ~300 lines
- [ ] All tests pass
- [ ] No broken references or missing helpers

## Note

After all split tickets (GPM-31 through GPM-34 and GPM-38) are completed, only TestIntegrationWorkflow and TestIntegrationInitValidation should remain in integration_test.go (~325 lines including helpers).