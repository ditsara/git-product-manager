---
id: GPM-74
title: "Replace raw SQL in ShouldSync with Bob query builder"
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

`ShouldSync()` in `internal/cache/sync.go` (~line 104) uses a raw `db.QueryRow()` call instead of the Bob query builder, inconsistent with the rest of the cache layer.

## Current code

```go
err = db.QueryRow("SELECT value FROM cache_metadata WHERE key = 'last_sync_timestamp'").Scan(&lastSyncStr)
```

## Target pattern

Use Bob Layer 1 SELECT, matching the style in `query.go`:

```go
querySQL, queryArgs, err := sqlite.Select(
    sm.Columns("value"),
    sm.From("cache_metadata"),
    sm.Where(sqlite.Quote("key").EQ(sqlite.Arg("last_sync_timestamp"))),
).Build(ctx)
err = db.QueryRowContext(ctx, querySQL, queryArgs...).Scan(&lastSyncStr)
```

## Acceptance Criteria

- No inline SQL string in `ShouldSync`
- Bob query builder used for the SELECT
- `make test` passes
