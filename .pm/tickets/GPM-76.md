---
id: GPM-76
title: "Replace raw SQL INSERT in SyncMilestones with Bob bulk-insert pattern"
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
