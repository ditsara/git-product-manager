---
id: GPM-35
title: "Add unit tests for cmd/pm/comment.go"
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

[Claude Haiku 4.5]

Add unit tests for comment.go which has 7 functions with complex logic not fully covered by integration tests.

## Functions to Test

- `selectCommentToAmend()` - Interactive menu selection
- `getCommentEditViaEditor()` - Editor selection fallback chain
- Author detection and fallback logic
- Comment file parsing and validation

## Create cmd/pm/comment_test.go with tests for:

- [ ] Interactive comment selection from multiple comments
- [ ] Editor selection (VISUAL, EDITOR, fallback chain)
- [ ] Author detection from git config
- [ ] Author override via flag
- [ ] Comment validation (empty check, etc.)

## Acceptance Criteria

- [ ] comment_test.go created with unit tests
- [ ] Coverage for all public functions
- [ ] All tests pass: `make test`
- [ ] No external dependencies needed (mock as needed)