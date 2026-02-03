---
id: GPM-17
title: "Implement cache metadata table for staleness tracking"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 3  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""  # Parent epic or story
depends_on: [GPM-10]  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: [cache, performance]  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-03T05:34:58Z"
updated_at: "2026-02-03T05:34:58Z"
---

# Description

[Sonnet 4.5]

Add automatic cache synchronization that detects when ticket files have been manually edited and rebuilds the cache transparently.

## Current Problem

**`pm list` reads directly from filesystem** (slow but always accurate):
- No caching means every `pm list` scans all ticket files
- Performance degrades with many tickets (100+ tickets)
- Users can't benefit from SQLite indexing and FTS

**Cache becomes stale when tickets are manually created/edited:**
- If user edits `.pm/tickets/GPM-123.md` directly, cache doesn't know
- `pm list` would show outdated information (if we used cache)
- Requires manual `pm reindex` command (which doesn't exist yet)

## Solution: Lazy Sync on Read

Before any query operation (`list`, `search`), automatically check if cache is stale and resync if needed:

1. **Store last sync timestamp** in a `cache_metadata` table
2. **Before each query:** Check if any `.md` file has `mtime > last_sync_timestamp`
3. **If stale:** Automatically rescan all tickets and update cache
4. **If fresh:** Use cached data (fast path)

This is transparent to users - the cache "just works" and is always correct.

## Implementation Steps

- [ ] Create migration `000002_add_cache_metadata.up.sql`:
  ```sql
  CREATE TABLE cache_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
  );
  INSERT INTO cache_metadata (key, value) VALUES ('last_sync_timestamp', '1970-01-01T00:00:00Z');
  ```
- [ ] Create `internal/cache/sync.go` with sync logic:
  - `ShouldSync(pmPath) (bool, error)`: Check if any ticket file mtime > last_sync_timestamp
  - `SyncCache(pmPath) error`: Scan `.pm/tickets/`, parse all tickets, update database
  - `UpdateSyncTimestamp(db) error`: Set last_sync_timestamp to now
- [ ] Update `cmd/pm/list.go`:
  - Before querying cache, check `ShouldSync()`
  - If true, call `SyncCache()` transparently
  - Then proceed with normal query
- [ ] Write tests:
  - Unit test for staleness detection (mock mtimes)
  - Integration test: create ticket manually, verify `pm list` auto-syncs and shows it
  - Test with no tickets, with 100 tickets, with nested directories

## Performance Considerations

- **Staleness check is fast:** Single filesystem scan to get mtimes (~1ms for 100 files)
- **Sync only when needed:** Fresh cache hits are instant (SQLite query)
- **Acceptable overhead:** Small price for correctness and no manual intervention
- **Future optimization:** Could batch updates or use filesystem watchers

## Edge Cases

- **First run after init:** last_sync_timestamp is epoch, triggers full sync
- **Clock skew:** Use filesystem mtime, not user's system clock
- **Deleted tickets:** Sync detects missing files and removes from cache
- **Concurrent access:** SQLite handles locking, last writer wins on timestamp

## Acceptance Criteria

- [ ] `pm list` automatically syncs cache when tickets are manually edited
- [ ] Creating ticket manually (without `pm new`) appears in next `pm list`
- [ ] Cache is never stale - always reflects current filesystem state
- [ ] No performance regression when cache is fresh (<10ms overhead)
- [ ] Migration adds `cache_metadata` table successfully

## Example Flow

```bash
# Initial state: cache is fresh
$ pm list  # Fast: uses cache, staleness check passes

# User manually edits ticket
$ vim .pm/tickets/GPM-123.md  # Updates mtime to now

# Next list detects staleness
$ pm list  # Detects mtime > last_sync, auto-syncs, then displays
# Output shows updated content from GPM-123.md

# Subsequent lists are fast again
$ pm list  # Fast: cache is fresh
```