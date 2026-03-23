---
id: GPM-68
title: "Refactor tickets db table using materialized paths"
type: story
status: done
priority: high
points: 5

parent: "GPM-61"
depends_on: []
blocks: [GPM-69]
related: []

labels: [database, refactoring]
assignee: ""
created_at: "2026-03-23T14:53:40Z"
updated_at: "2026-03-23T15:05:31Z"
---

## Dev Readiness Evaluation

**Result: ✅ Ready to implement**

### Files that need changes

| File | Change needed |
|------|--------------|
| `internal/migrations/000005_add_path_column.up.sql` | New — `ALTER TABLE tickets ADD COLUMN path TEXT` |
| `internal/migrations/000005_add_path_column.down.sql` | New — `ALTER TABLE tickets DROP COLUMN path` |
| `internal/cache/sync.go` — `ticketData` struct | Add `path string` field |
| `internal/cache/sync.go` — `im.Into("tickets", ...)` | Add `"path"` to column list |
| `internal/cache/sync.go` — `im.Values(sqlite.Arg(...))` | Add `t.path` in matching position |
| `internal/cache/migrate_test.go` — `expectedColumns` | Add `"path": false` (for correctness; test won't break without it) |

### Tests analysis

**`internal/cache/migrate_test.go`**
The `TestRunMigrations` test uses `PRAGMA table_info(tickets)` and validates
that every key in `expectedColumns` exists in the table. Crucially, it does
**not** fail on extra/unexpected columns — so adding `path` won't break the
test. However, `"path": false` should be added to `expectedColumns` so the test
actually verifies the new column is present.

**`internal/cache/sync_test.go`**
Queries use explicit column lists (`SELECT id, title, type, status FROM
tickets`). Adding a column has no effect.

**All integration tests (`integration_*_test.go`)**
Run `pm list` and assert on output lines/formatting. None inspect the raw SQL
or column schema. Unaffected.

**`cmd/pm/list_test.go`**
Tests query building logic using explicit column names. Unaffected.

### Raw SQL compatibility check

All SELECT queries across the codebase use **explicit column lists** (not
`SELECT *`). None will pick up or be broken by the new column:

| File | SQL fragment | Safe? |
|------|-------------|-------|
| `cmd/pm/list.go` | `SELECT id, title, type, status, ...` | ✅ |
| `cmd/pm/blocked.go` | `SELECT DISTINCT t.id, t.title, t.status, ...` | ✅ |
| `internal/cache/sync.go` | `SELECT value FROM cache_metadata WHERE ...` | ✅ (different table) |

The only INSERT that touches the `tickets` table is the bulk insert in
`sync.go` — it uses explicit column names, so it must be updated to include
`path`.

### Down migration note

Both SQLite drivers in use (`mattn/go-sqlite3 v1.14.22` bundling SQLite ~3.45,
and `modernc.org/sqlite v1.44.3`) support `ALTER TABLE DROP COLUMN` (requires
SQLite ≥ 3.35.0, released 2021). The down migration can use `ALTER TABLE
tickets DROP COLUMN path` safely.

**[Claude Sonnet 4.6]**

Add a `path` column to the `tickets` cache table using the Materialized Path
pattern. This encodes each ticket's full ancestor chain directly in the row,
eliminating the need for recursive CTEs in subtree queries.

## Background

The `pm list --parent X --all` flag combination currently requires a `WITH
RECURSIVE` CTE to walk the ticket hierarchy. This prevents the query from being
expressed via Bob's ORM query builder. The Materialized Path pattern solves
this at the data layer so the query layer can remain clean.

## What is Materialized Path?

Each row stores its full ancestor chain as a slash-delimited string:

| ticket | parent | path                |
|--------|--------|---------------------| | GPM-1  | (none) | `GPM-1`
| | GPM-2  | GPM-1  | `GPM-1/GPM-2`       | | GPM-3  | GPM-2  |
`GPM-1/GPM-2/GPM-3` |

This enables subtree queries via `WHERE path LIKE 'GPM-1/%'` — no recursive CTE
needed.

## Implementation Steps

- [ ] Create `internal/migrations/000005_add_path_column.up.sql`: `ALTER TABLE
  tickets ADD COLUMN path TEXT`
- [ ] Create `internal/migrations/000005_add_path_column.down.sql`: `ALTER
  TABLE tickets DROP COLUMN path`
- [ ] In `sync.go` `SyncCache`: after collecting all `ticketData` into a slice,
  build a `map[string]ticketData` by ID
- [ ] Implement `buildPath(id string, byID map[string]ticketData, visited
  map[string]bool) string`:
  - If no parent: return `id`
  - If parent not found in map (orphan): return `id` (fallback)
  - If cycle detected (visited): return `id` (safety fallback)
  - Otherwise: return `buildPath(parent, ...) + "/" + id`
- [ ] Compute path for each ticket and store in `ticketData.path`
- [ ] Add `path` to the bulk INSERT column list and values in `SyncCache`
- [ ] Add `path TEXT` field to the local `ticketData` struct in `sync.go`

## Edge Cases

- **Root ticket (no parent):** `path = id` (e.g., `"GPM-1"`)
- **Orphan ticket (parent ID not in current ticket set):** `path = id` —
  graceful fallback, no crash
- **Cycle in parent chain:** Use a `visited` set in `buildPath`; if a cycle is
  detected, fall back to `id`
- **Empty tickets directory:** No paths to compute; migration still runs
  cleanly
- **Existing cache:** Migration adds the column as `NULL`; next sync will
  populate it

## Acceptance Criteria

- [ ] Migration 000005 up/down files exist and apply cleanly
- [ ] `SyncCache` populates `path` for all tickets
- [ ] Root tickets have `path = id`
- [ ] Child tickets have `path = "parent_path/id"`
- [ ] Orphan tickets (missing parent) fall back to `path = id`
- [ ] No panic or error on cycles in parent chain
- [ ] All existing tests pass without modification

