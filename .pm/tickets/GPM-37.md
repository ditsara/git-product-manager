---
id: GPM-37
title: "Add unit tests for cmd/pm/init.go and cmd/pm/list.go"
type: task
status: backlog
priority: medium
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

[Claude Opus 4.6]

Add unit tests for init.go and list.go which have setup/filtering logic worth isolating.

## init.go Tests (6 functions)

- [ ] Config file creation with defaults
- [ ] Workflow template generation
- [ ] Directory structure setup
- [ ] Migration running
- [ ] Prefix uppercasing

## list.go Tests (2 functions)

- [ ] Filtering by status
- [ ] Filtering by assignee
- [ ] Multiple filter combinations (AND logic)
- [ ] Parent filtering (hierarchical)
- [ ] Sorting and column formatting

**Note:** List-related *integration* tests (TestIntegrationCacheSync, TestHierarchicalFiltering) are covered by GPM-38. This ticket focuses on *unit* tests for list.go's internal logic.

## Implementation Strategy

- Create separate `init_test.go` and `list_test.go` files
- Mock filesystem where needed
- Use t.TempDir() for isolated test environments

## Acceptance Criteria

- [ ] init_test.go with tests for setup logic
- [ ] list_test.go with tests for filtering/formatting
- [ ] All tests pass: `make test`
- [ ] No file I/O outside of t.TempDir()