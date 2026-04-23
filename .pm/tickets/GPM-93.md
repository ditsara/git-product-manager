---
assignee: ""
blocks: []
created_at: "2026-04-23T16:39:46Z"
depends_on: []
id: GPM-93
labels: []
parent: ""
points: 0
priority: high
related:
    - GPM-94
status: backlog
title: Fix syncTickets bulk INSERT SQLite parameter limit crash
type: bug
updated_at: "2026-04-23T16:40:52Z"
---

# Description

`syncTickets` (and `syncRelationships`, `syncComments`) build a single bulk `INSERT INTO tickets VALUES (?,?,...)`
statement with every ticket packed into one statement. SQLite's `SQLITE_MAX_VARIABLE_NUMBER` is 32,766. With 12
columns per ticket, `SyncCache` crashes with a SQLite error at **~2,731 tickets** — every subsequent command
that triggers a sync fails permanently until `.cache.db` is deleted.

## Impact

Hard failure (not degraded performance). Any project that reaches ~2,700 tickets has a broken `pm`.

## Recommended Fix

Paginate the bulk inserts in batches of N rows (e.g. 200) so the max bind-parameter count stays well under the
limit. All rows still land in the same transaction, so atomicity is preserved.

```
const insertBatchSize = 200 // 200 × 12 cols = 2400 params, well under 32766

for i := 0; i < len(tickets); i += insertBatchSize {
    end := min(i+insertBatchSize, len(tickets))
    batch := tickets[i:end]
    // build and exec INSERT for batch
}
```

Same pattern applies to `syncComments` (4 cols → safe at 8,000 rows/batch) and `syncRelationships`
(3 cols → safe at 10,000 rows/batch), but standardizing all three on 200-row batches keeps the code uniform.

## Acceptance Criteria

- `SyncCache` succeeds with 3,000+ ticket files without error
- All rows land in the same transaction (no partial syncs on failure)
- Existing tests continue to pass
- A test or benchmark with >2,731 tickets confirms the fix (or a unit test with a stubbed limit)
