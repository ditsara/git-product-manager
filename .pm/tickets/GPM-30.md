---
id: GPM-30
title: "Test infrastructure cleanup and expansion"
type: epic
status: backlog
priority: high
points: 0

parent: ""
depends_on: []
blocks: []
related: []

labels: [testing, refactoring, quality]
assignee: ""
created_at: "2026-02-08T08:16:30Z"
updated_at: "2026-02-08T08:16:30Z"
---

# Description

[Claude Opus 4.6]

Expand and reorganize test coverage to improve maintainability and reduce duplication. Currently, integration_test.go is 1,584 lines with 26 tests covering multiple features. Several command implementations lack unit tests despite containing complex logic.

## Objectives

1. **Split integration tests by feature domain** to reduce cognitive load and improve navigation
2. **Add unit tests** for complex command implementations (comment, edit, init, list)
3. **Maintain comprehensive coverage** with integration tests covering end-to-end workflows

## Current State

- **integration_test.go**: 1,584 lines, 26 tests (covers: comments, assign, show, history, list, init, workflow)
- **Missing unit tests**: comment.go (7 functions), edit.go (3), init.go (6), list.go (2)
- **Adequate coverage**: assign.go, move.go (thin wrappers), main.go (boilerplate)

## Child Tickets

**Integration Test Splits:**
- GPM-31: Split comment tests into `integration_comment_test.go` (9 tests)
- GPM-32: Split assign tests into `integration_assign_test.go` (5 tests)
- GPM-33: Split show tests into `integration_show_test.go` (5 tests)
- GPM-34: Split history tests into `integration_history_test.go` (3 tests)
- GPM-38: Split list/cache tests into `integration_list_test.go` (2 tests, many subtests)

**Remaining in integration_test.go:**
- TestIntegrationWorkflow — Core end-to-end happy path (init → new → list → show → move → edit)
- TestIntegrationInitValidation — Init validation (prefix requirements, uppercase conversion)
- Helper functions: buildPMBinary, getProjectRoot, runPM, initGitRepo, initWorkspace

**Unit Test Additions (High Priority):**
- GPM-35: Add unit tests for `cmd/pm/comment.go` (editor, author logic)
- GPM-36: Add unit tests for `cmd/pm/edit.go` (field parsing, validation)

**Unit Test Additions (Medium Priority):**
- GPM-37: Add unit tests for `cmd/pm/init.go` and `cmd/pm/list.go`

## Benefits

- Each integration test file ~200-300 lines (vs. 1,584 currently)
- Tests grouped by feature domain for faster navigation
- Complex logic isolated and tested at unit level
- Reduced debugging time when tests fail