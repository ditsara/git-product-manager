---
assignee: ""
blocks: []
created_at: "2026-04-23T16:40:03Z"
depends_on: []
id: GPM-94
labels: []
parent: ""
points: 0
priority: low
related:
    - GPM-93
status: backlog
title: Optimize has_children correlated subquery for large ticket sets
type: task
updated_at: "2026-04-23T16:40:52Z"
---

# Description

`ListTickets` emits a correlated subquery per result row to compute `has_children`:

```sql
CASE WHEN EXISTS(
  SELECT 1 FROM tickets AS t WHERE UPPER(t.parent) = UPPER(tickets.id)
) THEN 1 ELSE 0 END
```

There is no index on `parent`, and `UPPER()` prevents SQLite from using one even if it existed.
At 10,000 tickets this adds ~200–600 ms to every `pm list` invocation.

## Hot-take approaches (pick one, discuss before implementing)

**Option A — Index + avoid UPPER():** Store `parent` as uppercase at write time (normalize in
`applyTicketFields` and `ticket.Normalize`). Add `CREATE INDEX idx_parent ON tickets(parent)`.
The subquery becomes `WHERE t.parent = tickets.id` and uses the index. Simple, no schema change beyond the index.

**Option B — Left join aggregate:** Replace the correlated subquery with a single LEFT JOIN + GROUP BY
or a window function, letting SQLite count children once across all rows instead of N separate EXISTS checks.
More SQL complexity, similar or better performance.

**Option C — Precompute in cache:** Add a boolean `has_children` column to the tickets table, populated
during `SyncCache`. Zero query cost; adds one extra pass during sync (already O(n) anyway for path building).

## Acceptance Criteria

- `pm list` with 10,000 tickets returns in under 100 ms
- No behavioral change to output
- Existing tests pass
