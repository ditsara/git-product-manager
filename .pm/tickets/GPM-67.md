---
assignee: ""
blocks: []
created_at: "2026-02-14T10:57:04Z"
depends_on: []
id: GPM-67
labels:
    - database
    - refactoring
    - technical-debt
parent: GPM-61
points: 2
priority: low
related:
    - GPM-64
status: done
title: Migrate blocked.go JOIN queries to Bob
type: task
updated_at: "2026-03-23T16:06:03Z"
---


# Description

**[Claude Sonnet 4.6]**

Migrate the two remaining raw SQL JOIN queries in `cmd/pm/blocked.go →
showTicketBlockedView()` to use Bob's `sqlite/sm` query builder, consistent
with the patterns established in GPM-69.

## Current State

**File:** `cmd/pm/blocked.go`
**Function:** `showTicketBlockedView()` — two raw SQL variables:

```sql
-- dependsOnQuery (~line 215)
SELECT r.to_ticket, t.title, t.status
FROM relationships r
JOIN tickets t ON t.id = r.to_ticket
WHERE r.from_ticket = ? AND r.relationship_type = 'depends-on'
ORDER BY r.to_ticket

-- blocksQuery (~line 254)
SELECT r.from_ticket, t.title, t.status
FROM relationships r
JOIN tickets t ON t.id = r.from_ticket
WHERE r.to_ticket = ? AND r.relationship_type = 'depends-on'
ORDER BY r.from_ticket
```

The simple ticket-exists SELECT on line ~192 was already migrated in GPM-62.

## Bob API — confirmed patterns (from GPM-69 work)

All expressions confirmed available in `github.com/stephenafamo/bob v0.42.0`:

| SQL | Bob |
|-----|-----|
| `FROM relationships r` | `sm.From("relationships AS r")` |
| `JOIN tickets t ON t.id = r.to_ticket` | `sm.InnerJoin("tickets AS t").OnEQ(sqlite.Quote("t", "id"), sqlite.Quote("r", "to_ticket"))` |
| `WHERE r.from_ticket = ?` | `sm.Where(sqlite.Quote("r", "from_ticket").EQ(sqlite.Arg(ticketID)))` |
| `AND r.relationship_type = 'depends-on'` | `sm.Where(sqlite.Quote("r", "relationship_type").EQ(sqlite.Arg("depends-on")))` — Bob ANDs multiple `Where` mods automatically |
| `ORDER BY r.to_ticket` | `sm.OrderBy(sqlite.Quote("r", "to_ticket"))` |
| `sqlite.Quote("t", "id")` | multi-arg form produces `"t"."id"` (table-qualified) |

**Important:** `sqlite.F(name, args...)` returns `mods.Moddable[*dialect.Function]` — call it
as a function `sqlite.F(...)()` to get the `*dialect.Function` value with `.EQ()` chaining.
Not needed for this ticket (no UPPER() calls), but noted for awareness.

## Bob translation

```go
// dependsOnQuery
q := sqlite.Select(
    sm.Columns(
        sqlite.Quote("r", "to_ticket"),
        sqlite.Quote("t", "title"),
        sqlite.Quote("t", "status"),
    ),
    sm.From("relationships AS r"),
    sm.InnerJoin("tickets AS t").OnEQ(
        sqlite.Quote("t", "id"),
        sqlite.Quote("r", "to_ticket"),
    ),
    sm.Where(sqlite.Quote("r", "from_ticket").EQ(sqlite.Arg(ticketID))),
    sm.Where(sqlite.Quote("r", "relationship_type").EQ(sqlite.Arg("depends-on"))),
    sm.OrderBy(sqlite.Quote("r", "to_ticket")),
)
querySQL, args, err := q.Build(context.Background())
// then: db.QueryContext(ctx, querySQL, args...)
```

The `blocksQuery` is symmetrical — swap `to_ticket`↔`from_ticket` and the `OnEQ` direction,
change `from_ticket = ?` to `to_ticket = ?`, order by `r.from_ticket`.

## Implementation Steps

- [x] Replace `dependsOnQuery` raw SQL variable with Bob `sqlite.Select(...)` builder
- [x] Replace `blocksQuery` raw SQL variable with Bob `sqlite.Select(...)` builder
- [x] Use `db.QueryContext` consistently (replace `db.Query`)
- [x] Run `go test ./...` to verify no regression

## Acceptance Criteria

- [x] No raw SQL strings in `showTicketBlockedView()`
- [x] Both queries use `sqlite.Select` + `sm.*` with no `sqlite.Raw()` fallbacks
- [x] `integration_blocked_test.go` passes unchanged
- [x] `go build ./...` succeeds

## Notes

- `showGlobalBlockedView()` is explicitly out of scope — its `GROUP_CONCAT` +
  dynamic `HAVING NOT IN` genuinely can't be expressed cleanly via Bob.
  That function should stay as raw SQL.
- No new files needed; only `cmd/pm/blocked.go` changes.
- No migration needed — this is command-layer code, not cache layer.
