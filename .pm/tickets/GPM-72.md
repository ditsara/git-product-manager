---
id: GPM-72
title: "Code Cleanup: Large Functions and Raw SQL"
type: epic
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""  # Parent epic (for nested epics)
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-03-26T01:45:52Z"
updated_at: "2026-03-26T01:45:52Z"
---

# Description

Code audit identified two categories of cleanup needed in the codebase.

## Large Functions (>200 lines)

| Ticket | File | Function | Lines |
|--------|------|----------|-------|
| GPM-73 | `internal/cache/sync.go` | `SyncCache` | ~243 (L161–403) |

Two integration test functions also exceed 200 lines (`TestHierarchicalFiltering` at ~296 lines, `TestIntegrationWorkflow` at ~215 lines) but are deferred — test readability is lower priority.

## Raw SQL Instead of Bob ORM (Layer 1)

| Ticket | File | Function | Operation | Issue |
|--------|------|----------|-----------|-------|
| GPM-74 | `internal/cache/sync.go` | `ShouldSync` | SELECT | `db.QueryRow()` with inline SQL |
| GPM-75 | `internal/cache/sync.go` | `SyncMilestones` | DELETE | `tx.Exec("DELETE FROM milestones")` |
| GPM-76 | `internal/cache/sync.go` | `SyncMilestones` | INSERT OR REPLACE | Raw parameterized SQL in loop |

One additional raw SQL instance (`showGlobalBlockedView` in `cmd/pm/blocked.go`) was identified but intentionally excluded — it uses `GROUP_CONCAT` + `HAVING` which Bob does not cleanly support, and has a documented justification.

## Acceptance Criteria

- All sub-tickets (GPM-73–GPM-76) resolved
- No regressions: `make test` passes
