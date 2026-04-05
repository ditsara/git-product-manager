---
assignee: ""
blocks: []
created_at: "2026-03-26T01:46:01Z"
depends_on: []
id: GPM-76
labels: []
parent: GPM-72
points: 0
priority: medium
related: []
status: done
title: Replace raw SQL INSERT in SyncMilestones with Bob bulk-insert pattern
type: task
updated_at: "2026-04-05T08:47:55Z"
---


# Description

`SyncMilestones()` in `internal/cache/sync.go` (~lines 464–468) uses a raw parameterized INSERT inside a loop, while the rest of `sync.go` uses Bob's bulk-insert pattern for tickets, comments, and relationships.

## Current code

```go
for _, m := range milestones {
    _, err := tx.Exec(
        `INSERT OR REPLACE INTO milestones (id, title, description, due_date, state, created_at, closed_at)
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
        m.ID, m.Title, m.Description, m.DueDate, m.State, m.CreatedAt, m.ClosedAt,
    )
    if err != nil {
        return fmt.Errorf(...)
    }
}
```

## Target pattern

Use the Bob bulk-insert pattern matching the ticket/comment/relationship inserts in `SyncCache` (L311–382):

```go
insertMilestones := sqlite.Insert(
    im.Into("milestones", "id", "title", "description", "due_date", "state", "created_at", "closed_at"),
    im.OrReplace(),
)
for _, m := range milestones {
    insertMilestones.Apply(im.Values(
        sqlite.Arg(m.ID), sqlite.Arg(m.Title), sqlite.Arg(m.Description),
        sqlite.Arg(m.DueDate), sqlite.Arg(m.State), sqlite.Arg(m.CreatedAt), sqlite.Arg(m.ClosedAt),
    ))
}
insertSQL, insertArgs, err := insertMilestones.Build(ctx)
if err != nil {
    return fmt.Errorf("building milestones insert: %w", err)
}
if _, err = tx.ExecContext(ctx, insertSQL, insertArgs...); err != nil {
    return fmt.Errorf("inserting milestones cache: %w", err)
}
```

## Notes

- GPM-75 (DELETE) and this ticket (INSERT) are both in `SyncMilestones` — they can be done together in one PR.
- `im.OrReplace()` is consistent with the existing `INSERT OR REPLACE` intent of the current code.

## Acceptance Criteria

- No inline SQL string for the INSERT in `SyncMilestones`
- Single bulk insert replaces the per-row loop
- `make test` passes
