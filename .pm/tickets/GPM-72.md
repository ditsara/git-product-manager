---
assignee: ""
blocks: []
created_at: "2026-03-26T01:45:52Z"
depends_on: []
id: GPM-72
labels: []
parent: ""
points: 0
priority: medium
related: []
status: done
title: 'Code Cleanup: Large Functions and Raw SQL'
type: epic
updated_at: "2026-04-05T08:51:46Z"
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

## Completion Notes

| Ticket | Status | Notes |
|--------|--------|-------|
| GPM-73 | ✅ done | `SyncCache` reduced from 243 → 59 lines; 5 helpers extracted |
| GPM-74 | ✅ done | `ShouldSync` SELECT now uses Bob `sm.Select` |
| GPM-75 | ✅ done | `SyncMilestones` DELETE now uses `clearTable` helper |
| GPM-76 | ✅ done | `SyncMilestones` INSERT loop replaced with Bob bulk insert |

`showGlobalBlockedView` in `cmd/pm/blocked.go` intentionally skipped — `GROUP_CONCAT` + `HAVING` with dynamic `NOT IN` is not cleanly expressible in Bob Layer 1; raw SQL with a documented comment is the correct call here.
