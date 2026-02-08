---
id: GPM-35
title: "Add unit tests for cmd/pm/comment.go"
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

[Claude Haiku 4.5]

Add unit tests for comment.go which has 7 functions with complex logic not fully covered by integration tests.

## Functions to Test

- `selectCommentToAmend()` - Interactive menu selection
- `getCommentEditViaEditor()` - Editor selection fallback chain
- Author detection and fallback logic
- Comment file parsing and validation

## Create cmd/pm/comment_test.go with tests for:

- [x] Interactive comment selection from multiple comments
- [x] Editor selection (VISUAL, EDITOR, fallback chain)
- [x] Author detection from git config
- [x] Author override via flag
- [x] Comment validation (empty check, etc.)

## Acceptance Criteria

- [x] comment_test.go created with unit tests
- [x] Coverage for all public functions
- [x] All tests pass: `make test`
- [x] No external dependencies needed (mock as needed)

## Implementation Notes

Created comprehensive unit tests in cmd/pm/comment_test.go with 8 test functions:
- TestSelectCommentToAmend: Tests interactive menu selection with valid/invalid input and cancellation
- TestGetEditorFallbackChain: Tests VISUAL/EDITOR env var precedence and fallback chain
- TestGetGitAuthor: Tests git config author detection and whitespace handling
- TestFilterCommentLines: Tests comment line removal logic with various inputs
- TestCommentValidation: Tests validation of comment bodies (empty, whitespace-only, etc.)
- TestCommentFileContent: Tests YAML front matter generation with author and timestamp fields
- TestCommentTimestampParsing: Tests RFC3339 timestamp parsing for comment amendments
- TestAuthorDetectionLogic: Tests author flag override logic and git config fallback

All tests passing. No external mocking needed - uses standard Go testing patterns.