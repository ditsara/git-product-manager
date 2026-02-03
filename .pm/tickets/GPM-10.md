---
id: GPM-10
title: "Lazy migration check on every command"
type: task
status: backlog
priority: high
points: 2

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""
depends_on: []
blocks: [GPM-9]
related: []

labels: [cache, migrations, reliability]
assignee: ""
created_at: "2026-02-03T03:52:16Z"
updated_at: "2026-02-03T03:52:16Z"
---

# Description

[Sonnet 4.5]

Ensure the cache database schema is always up-to-date by checking migrations before every command that uses the cache.

## Current Problem

`pm init` runs migrations to create the database, but:
- If `.cache.db` is deleted, subsequent commands fail with "no such table" errors
- If schema is outdated (new migration added), commands fail or behave incorrectly
- Users must manually detect and fix schema issues

## Solution

Add lazy migration check at the start of every cache operation:

```go
// Before any cache query:
func ensureCacheReady(pmPath string) error {
    dbPath := filepath.Join(pmPath, ".cache.db")
    
    // Check if database exists
    if _, err := os.Stat(dbPath); os.IsNotExist(err) {
        // Database missing - run migrations to create it
        if err := cache.RunMigrations(dbPath, migrationPath); err != nil {
            return fmt.Errorf("failed to initialize cache: %w", err)
        }
    } else {
        // Database exists - check if migrations needed
        if err := cache.RunMigrations(dbPath, migrationPath); err != nil {
            return fmt.Errorf("failed to update cache schema: %w", err)
        }
    }
    
    return nil
}
```

## Implementation Steps

- [ ] Create `internal/cache/ensure.go` with `EnsureCacheReady(pmPath) error`
- [ ] Call `EnsureCacheReady()` in `list.go` before opening database
- [ ] Call `EnsureCacheReady()` in `show.go`, `search.go`, and other cache users
- [ ] Update `cache.RunMigrations()` to be idempotent (safe to call multiple times)
- [ ] Add tests: missing db, outdated schema, already current

## Performance Considerations

- `golang-migrate` is smart: if schema is current, it's a fast no-op
- Small overhead (~few milliseconds) is acceptable for reliability
- Could add caching if needed: check schema version in memory

## Acceptance Criteria

- [ ] Deleting `.cache.db` doesn't break any command (auto-recreates)
- [ ] Adding new migration automatically updates existing databases
- [ ] No performance regression for normal usage (current schema)
- [ ] Works correctly with concurrent commands (database locking)

## Related

Blocks GPM-9 (auto-recovery) which depends on robust migration handling.
