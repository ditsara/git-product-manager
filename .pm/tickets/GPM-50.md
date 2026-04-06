---
assignee: ""
blocks: []
created_at: "2026-02-08T14:55:20Z"
depends_on:
    - GPM-45
id: GPM-50
labels:
    - relationships
    - cli
    - filtering
parent: GPM-2
points: 3
priority: medium
related: []
status: done
title: Enhance pm list with relationship-aware filtering
type: task
updated_at: "2026-04-06T14:48:33Z"
---



# Description

## Current State

The `relationships` table (migration 000004) already exists and is populated by cache sync:

- `depends-on` rows: `from_ticket` depends on `to_ticket`
- `blocks` rows: `from_ticket` blocks `to_ticket`

`related` rows are **not yet synced** — `sync.go` only iterates `DependsOn` and `Blocks` arrays.

`pm blocked` already queries this table directly. The remaining work is syncing `related`, then wiring all three relationship types into `pm list` via new filter flags.

## Filter Flags to Add

```bash
pm list --depends-on <id>   # Show tickets that depend on <id>
pm list --blocks <id>       # Show tickets that <id> blocks (i.e., depend on it)
pm list --related <id>      # Show tickets with any relationship to <id>
```

## Implementation: relationships table queries

Each flag maps to a JOIN on the `relationships` table:

**`--depends-on <id>`** — tickets whose `from_ticket` has a `depends-on` row pointing to `<id>`:
```sql
JOIN relationships r ON r.from_ticket = t.id
  AND r.to_ticket = <id>
  AND r.relationship_type = 'depends-on'
```

**`--blocks <id>`** — tickets that `<id>` depends on (reverse lookup, `<id>` is the `from_ticket`):
```sql
JOIN relationships r ON r.to_ticket = t.id
  AND r.from_ticket = <id>
  AND r.relationship_type = 'depends-on'
```

> Note: `blocks` rows are also stored during sync (from the ticket's `blocks:` YAML field). Those can also be queried as `r.from_ticket = <id> AND r.relationship_type = 'blocks'`.

**`--related <id>`** — tickets with any relationship to/from `<id>`:
```sql
JOIN relationships r ON (r.from_ticket = t.id AND r.to_ticket = <id>)
                      OR (r.to_ticket = t.id AND r.from_ticket = <id>)
```

## Changes Required

- In `internal/cache/sync.go` (`scanTicketFiles`): add a loop over `t.Related` to append `related` rows into the relationships slice (alongside the existing `depends-on` and `blocks` loops)
- Add `DependsOn`, `Blocks`, `Related` string fields to `cache.ListOptions`
- Update `cache.ListTickets` to JOIN `relationships` when any of those fields is set
- Add `--depends-on`, `--blocks`, `--related` flags to `listCmd` in `cmd/pm/list.go`
- Wire flag values into `ListOptions` before calling `cache.ListTickets`
- Add shell completion for the new flags (reuse `completeTicketIDs`)
- Add tests in `internal/cache/query_test.go` covering each filter and combinations

## Acceptance Criteria

- `related` rows are populated in the `relationships` table during cache sync
- `--depends-on` returns all tickets that depend on the target
- `--blocks` returns all tickets blocked by (i.e., waiting on) the target
- `--related` returns tickets with any relationship to the target (bidirectional)
- All three flags combine with AND logic alongside existing filters
- Shell completion works for new flags
- All tests pass
