---
assignee: ""
blocks: []
created_at: "2026-03-26T01:46:01Z"
depends_on: []
id: GPM-74
labels: []
parent: GPM-72
points: 0
priority: medium
related: []
status: done
title: Replace raw SQL in ShouldSync with Bob query builder
type: task
updated_at: "2026-04-05T08:47:55Z"
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
