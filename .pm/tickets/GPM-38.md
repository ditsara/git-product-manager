---
id: GPM-38
title: "Split list and cache tests into integration_list_test.go"
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

- [ ] Create `integration_list_test.go`
- [ ] Move TestIntegrationCacheSync from integration_test.go
- [ ] Move TestHierarchicalFiltering from integration_test.go
- [ ] Ensure helper functions (initWorkspace, runPM, buildPMBinary) are accessible
- [ ] Remove moved tests from integration_test.go
- [ ] Verify all tests still pass: `make test`

## Acceptance Criteria

- [ ] New file has ~415 lines with 2 top-level tests (many subtests)
- [ ] integration_test.go reduced to ~325 lines with 2 core tests + helpers
- [ ] All tests pass
- [ ] No broken references or missing helpers
