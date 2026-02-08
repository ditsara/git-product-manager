---
id: GPM-38
title: "Split list and cache tests into integration_list_test.go"
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
created_at: "2026-02-08T08:24:01Z"
updated_at: "2026-02-08T08:24:01Z"
---

# Description

[Claude Opus 4.6]

Extract list and cache sync integration tests from integration_test.go into a new dedicated test file. These tests cover `pm list` filtering, hierarchical parent/child behavior, and cache auto-sync when tickets are created or modified outside the CLI.

## Tests to Move

- TestIntegrationCacheSync (line 368, ~120 lines) — Tests cache auto-sync with manual ticket creation/modification
- TestHierarchicalFiltering (line 489, ~295 lines) — Tests `--parent`, `--all`, combined filters, orphaned tickets

## What Stays in integration_test.go

After all split tickets (GPM-31 through GPM-34 and GPM-38) are completed, the following core tests remain in integration_test.go as the baseline sanity check:

- TestIntegrationWorkflow — End-to-end happy path: init → new → list → show → move → edit → gap handling
- TestIntegrationInitValidation — Init without prefix, lowercase prefix conversion

These two tests validate the foundational workflow and should stay together as the "core" integration test.

## Implementation Steps

- [x] Create `integration_list_test.go`
- [x] Move TestIntegrationCacheSync from integration_test.go
- [x] Move TestHierarchicalFiltering from integration_test.go
- [x] Ensure helper functions (initWorkspace, runPM, buildPMBinary) are accessible
- [x] Remove moved tests from integration_test.go
- [x] Verify all tests still pass: `make test`

## Acceptance Criteria

- [x] New file has ~415 lines with 2 top-level tests (many subtests)
- [x] integration_test.go reduced to ~325 lines with 2 core tests + helpers
- [x] All tests pass
- [x] No broken references or missing helpers

## Implementation Notes

- Fixed initial file creation issue with duplicate "package main" declarations
- Recreated integration_list_test.go with single clean package declaration
- All 26 integration tests accounted for across the five split files:
  - integration_comment_test.go: 9 tests (330 lines)
  - integration_assign_test.go: 5 tests (157 lines)
  - integration_show_test.go: 5 tests (206 lines)
  - integration_history_test.go: 3 tests (159 lines)
  - integration_list_test.go: 2 tests with many subtests (~415 lines) ✅ NEW
  - integration_test.go: 2 core tests (~325 lines)
- All tests pass with `make test`
