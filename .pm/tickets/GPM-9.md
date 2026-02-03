---
id: GPM-9
title: "Auto-recovery on database errors"
type: task
status: backlog
priority: high
points: 3

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""
depends_on: [GPM-10]
blocks: []
related: [GPM-11]

labels: [cache, reliability, ux]
assignee: ""
created_at: "2026-02-03T03:52:15Z"
updated_at: "2026-02-03T03:52:15Z"
---

# Description

[Sonnet 4.5]

Make the system self-healing by automatically recovering from database errors instead of failing with cryptic error messages.

## Current Problem

When the cache database is missing or corrupted, commands fail with unhelpful errors:
```
Error: no such table: tickets
Error: database is locked
Error: file is not a database
```

Users must manually diagnose the issue and run recovery steps (delete cache, reinitialize, etc.).

## Solution

Detect common database errors and automatically rebuild the cache:

1. **Error Detection**: Catch specific database errors:
   - `no such table` → Missing schema
   - `database is locked` → Concurrent access issue
   - `file is not a database` → Corrupted file
   - File doesn't exist → Missing cache

2. **Auto-Recovery Strategy**:
   - Log warning to user: `⚠ Cache corrupted, rebuilding...`
   - Delete corrupted `.cache.db` file
   - Run migrations to create fresh database
   - Sync from filesystem (call `cache.SyncCache()`)
   - Continue with original command
   - Log success: `✓ Cache rebuilt successfully`

3. **Implementation Location**:
   - Create `internal/cache/recovery.go` with `RecoverFromError(err, pmPath) error`
   - Wrap all database operations in recovery logic
   - Add to `list.go`, `show.go`, and any other cache-using commands

## Acceptance Criteria

- [ ] Deleting `.cache.db` doesn't break `pm list` (auto-recreates)
- [ ] Corrupting `.cache.db` doesn't break commands (detects and rebuilds)
- [ ] User sees helpful messages during recovery (not silent failures)
- [ ] Recovery is attempted at most once per command (avoid infinite loops)
- [ ] Original command completes successfully after recovery

## Related

Depends on GPM-10 (lazy migration check) for robust initialization.
Works alongside GPM-11 (pm repair) for explicit manual recovery.
