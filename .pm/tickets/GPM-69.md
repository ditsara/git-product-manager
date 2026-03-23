---
id: GPM-69
title: "remove SQL from list.go using ORM"
type: story
status: backlog
priority: high
points: 5

parent: "GPM-61"
depends_on: [GPM-68]
blocks: []
related: []

labels: [database, refactoring]
assignee: ""
created_at: "2026-03-23T14:54:00Z"
updated_at: "2026-03-23T14:54:00Z"
---

# Description

**[Claude Sonnet 4.6]**

Remove all raw SQL strings, `strings.Contains` WHERE-detection hacks, and manual `rows.Scan` from `cmd/pm/list.go` by introducing a typed `CachedTicket` model and a `ListTickets(db, opts)` function in the cache layer, using Bob's `sqlite/sm` query builder throughout.

Depends on GPM-68 (materialized path column), which makes the previously-problematic subtree query expressible via Bob.

## Current Problems in `list.go`

1. **String concatenation for WHERE clauses** — checks `strings.Contains(query, "WHERE")` then appends `" AND ..."` or `" WHERE ..."` manually
2. **Four separate raw SQL string variables** for different query paths
3. **Manual `rows.Scan`** into bare local variables (`var id, title, ticketType, status string; var hasChildren int`)
4. **SQL logic in the command layer** — belongs in `internal/cache/`

## New Structures (`internal/cache/query.go`)

```go
// CachedTicket is a row from the tickets cache table, for pm list display.
type CachedTicket struct {
    ID          string `db:"id"`
    Title       string `db:"title"`
    Type        string `db:"type"`
    Status      string `db:"status"`
    HasChildren bool   `db:"has_children"`
}

// ListOptions controls filtering in ListTickets.
type ListOptions struct {
    ParentFilter  string   // show children of this ticket ID
    Subtree       bool     // if true + ParentFilter: all descendants via path LIKE
    IncludeStates []string // whitelist of statuses
    ExcludeStates []string // blacklist of statuses
}
```

## `ListTickets` Query Logic

All four query paths use Bob `sqlite/sm`; no raw SQL strings:

| Case | Base WHERE clause |
|------|-------------------|
| `--parent X --all` (subtree) | Fetch parent's `path`, then `path LIKE '{parent_path}/%'` |
| `--parent X` (direct children) | `UPPER(parent) = UPPER(?)` |
| default (top-level) | `parent IS NULL OR parent = ''` |
| `--all` | *(no WHERE)* |

Status filtering appended dynamically via `sm.Where(sqlite.Quote("status").In(...))` or `.NotIn(...)`.

The `has_children` computed column is expressed as a Bob raw expression constant reused across all paths.

Scanning uses `stephenafamo/scan` (already a dependency at v0.7.0) to populate `[]CachedTicket`.

## Implementation Steps

- [ ] Create `internal/cache/query.go`
- [ ] Define `CachedTicket` struct with `db:` tags
- [ ] Define `ListOptions` struct
- [ ] Implement `ListTickets(db *sql.DB, opts ListOptions) ([]CachedTicket, error)`:
  - [ ] Subtree case: query parent's `path`, then `sm.Where(path LIKE ...)`
  - [ ] Direct children case: `sm.Where(UPPER(parent) = UPPER(?))`
  - [ ] Top-level case: `sm.Where(parent IS NULL OR parent = '')`
  - [ ] All case: no WHERE predicate
  - [ ] Status include/exclude: append `sm.Where` for `IN`/`NOT IN` dynamically
  - [ ] Use `stephenafamo/scan` to scan rows into `[]CachedTicket`
- [ ] Update `cmd/pm/list.go`:
  - [ ] Remove raw SQL variables and `strings.Contains` logic
  - [ ] Build `cache.ListOptions` from cobra flags
  - [ ] Call `cache.ListTickets(db, opts)`
  - [ ] Replace `rows.Next` loop with range over `[]CachedTicket`

## Acceptance Criteria

- [ ] No raw SQL strings in `list.go`
- [ ] No `strings.Contains` or manual `AND`/`WHERE` appending in `list.go`
- [ ] `CachedTicket` and `ListOptions` defined in `internal/cache/query.go`
- [ ] All four filter modes work correctly (subtree, direct children, top-level, all)
- [ ] Status flags (`--status`, `--all`, `--completed`, `--active`, `--incomplete`) behave identically
- [ ] `has_children` indicator still shown in output
- [ ] All existing integration tests pass without modification

