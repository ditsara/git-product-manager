---
assignee: ""
blocks: []
created_at: "2026-03-26T01:46:01Z"
depends_on: []
id: GPM-75
labels: []
parent: GPM-72
points: 0
priority: medium
related: []
status: done
title: Replace raw SQL DELETE in SyncMilestones with Bob query builder
type: task
updated_at: "2026-04-05T08:47:55Z"
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
