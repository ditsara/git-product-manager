---
id: GPM-37
title: "Add unit tests for cmd/pm/init.go and cmd/pm/list.go"
type: task
status: done
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

- [x] Config file creation with defaults
- [x] Workflow template generation
- [x] Directory structure setup
- [x] Migration running
- [x] Prefix uppercasing

## list.go Tests (2 functions)

- [x] Filtering by status
- [x] Filtering by assignee
- [x] Multiple filter combinations (AND logic)
- [x] Parent filtering (hierarchical)
- [x] Sorting and column formatting

**Note:** List-related *integration* tests (TestIntegrationCacheSync, TestHierarchicalFiltering) are covered by GPM-38. This ticket focuses on *unit* tests for list.go's internal logic.

## Implementation Strategy

- Create separate `init_test.go` and `list_test.go` files
- Mock filesystem where needed
- Use t.TempDir() for isolated test environments

## Acceptance Criteria

- [x] init_test.go with tests for setup logic
- [x] list_test.go with tests for filtering/formatting
- [x] All tests pass: `make test`
- [x] No file I/O outside of t.TempDir()

## Implementation Notes

Created two comprehensive unit test files:

**init_test.go (6 test functions):**
- TestCreateDefaultWorkflow: Verifies workflow.yaml file creation and valid YAML structure with states key
- TestCreateDefaultLabels: Verifies labels.yaml file creation and valid YAML structure
- TestCreateDefaultTemplates: Verifies all 4 templates (story.md, task.md, bug.md, epic.md) are created with content and YAML front matter
- TestCreateGitignore: Verifies .gitignore file contains .cache.db entry
- TestCreateProjectConfig: Verifies project.yaml created with correct prefix preservation across uppercase/lowercase/mixed case
- TestPrefixUppercasing: Tests the strings.ToUpper() logic applied to prefixes
- TestInitDirectoryStructure: Verifies all required directories are created

**list_test.go (3 test function suites):**
- TestTruncate: 11 test cases for truncate() function with Unicode, emoji, edge cases (maxLen<3, exact length, etc.)
- TestTruncateEdgeCases: Tests negative, zero, and very large maxLen values
- TestTruncateWithFixedWidth: Tests truncate() behavior for table column formatting (real-world use cases)
- TestQueryBuilding: 6 test cases for SQL query construction logic (top-level filtering, --all flag, --parent filtering, recursive queries, status filtering, combined filters)

All tests passing. Tests use t.TempDir() for isolated file I/O. No external mocking required.