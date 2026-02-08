---
id: GPM-33
title: "Split show tests into integration_show_test.go"
type: task
status: backlog
priority: medium
points: 1

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

Extract all pm show command tests from integration_test.go into a dedicated test file.

## Tests to Move

- TestShowWithComments
- TestShowNoComments
- TestShowWithNoCommentsFlag
- TestShowCommentChronologicalOrder
- TestShowCaseInsensitiveWithComments

## Implementation Steps

- [ ] Create `integration_show_test.go`
- [ ] Move all show-related tests from integration_test.go
- [ ] Ensure helper functions are accessible
- [ ] Remove moved tests from integration_test.go
- [ ] Verify all tests pass: `make test`

## Acceptance Criteria

- [ ] New file has ~200 lines with 5 show tests
- [ ] All tests pass
- [ ] No broken references