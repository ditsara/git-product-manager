---
id: GPM-75
title: "Replace raw SQL DELETE in SyncMilestones with Bob query builder"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-72"  # Parent epic or story
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-03-26T01:46:01Z"
updated_at: "2026-03-26T01:46:01Z"
---

# Description

`SyncMilestones()` in `internal/cache/sync.go` (~line 449) uses a raw `tx.Exec()` call for its DELETE, despite the `clearTable()` helper already existing for exactly this purpose.

## Current code

```go
if _, err := tx.Exec("DELETE FROM milestones"); err != nil {
    return fmt.Errorf("clearing milestones cache: %w", err)
}
```

## Target pattern

Use the existing `clearTable` helper (which already uses Bob internally), consistent with how `SyncCache` clears tickets, comments, and relationships:

```go
if err := clearTable(ctx, tx, "milestones"); err != nil {
    return fmt.Errorf("clearing milestones cache: %w", err)
}
```

If `clearTable` isn't the right fit, use Bob directly:

```go
delSQL, delArgs, err := sqlite.Delete(dm.From("milestones")).Build(ctx)
if _, err = tx.ExecContext(ctx, delSQL, delArgs...); err != nil {
    return fmt.Errorf("clearing milestones cache: %w", err)
}
```

## Acceptance Criteria

- No inline SQL string for the DELETE in `SyncMilestones`
- `make test` passes
