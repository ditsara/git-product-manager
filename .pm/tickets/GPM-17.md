---
assignee: ""
blocks: []
created_at: "2026-02-03T05:34:58Z"
depends_on: []
id: GPM-17
labels:
    - cache
    - performance
parent: ""
points: 3
priority: medium
related: []
status: done
title: Implement cache metadata table for staleness tracking
type: task
updated_at: "2026-04-12T12:11:04Z"
---

## Implementation (completed)

All steps below were implemented as part of earlier work.

### Migration (000002_add_cache_metadata.up.sql) ✅
`cache_metadata` table with `key TEXT PRIMARY KEY, value TEXT NOT NULL`.
Seeded with `last_sync_timestamp = '1970-01-01T00:00:00Z'` to force first sync.

### internal/cache/sync.go ✅
Single file handling all sync logic:
- `SyncCache(pmPath) error` — full transactional rebuild: clears tickets/comments/relationships, scans all `.md` files, computes materialized paths, bulk-inserts, updates timestamp. Non-fatal milestone sync follows.
- `updateSyncTimestamp(tx)` — INSERT OR REPLACE into cache_metadata.
- `SyncMilestones(pmPath)` — separate milestone sync used after ticket transaction commits.
- Helper functions: `scanTicketFiles`, `scanCommentDirs`, `buildPath`, `clearTable`, `syncTickets`, `syncComments`, `syncRelationships`.

### Lazy sync in all read commands ✅
Pattern `ShouldSync → SyncCache` is in: `list`, `search`, `tree`, `blocked`, `comment`, `sync`.

## Acceptance Criteria

- [x] `pm list` automatically syncs cache when tickets are manually edited
- [x] Creating ticket manually (without `pm new`) appears in next `pm list`
- [x] Cache is never stale — always reflects current filesystem state
- [x] No performance regression when cache is fresh (<10ms overhead)
- [x] Migration adds `cache_metadata` table successfully