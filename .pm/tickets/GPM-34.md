---
id: GPM-34
title: "Split history tests into integration_history_test.go"
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

Extract all pm history command tests from integration_test.go into a dedicated test file.

## Tests to Move

- TestHistorySingleChange
- TestHistoryNoGitRepo
- TestHistoryTicketNotCommitted

## Implementation Steps

- [ ] Create `integration_history_test.go`
- [ ] Move all history-related tests from integration_test.go
- [ ] Ensure initGitRepo helper is accessible
- [ ] Remove moved tests from integration_test.go
- [ ] Verify all tests pass: `make test`

## Acceptance Criteria

- [ ] New file has ~150 lines with 3 history tests
- [ ] All tests pass
- [ ] No broken references