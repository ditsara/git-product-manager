---
id: GPM-27
title: "Auto-update updated_at from git history with mtime fallback"
type: story
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: GPM-44  # Parent epic or story
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-04T04:34:12Z"
updated_at: "2026-02-04T04:34:12Z"
---

**[Sonnet 4.5]** Implement incremental git-based timestamp computation for `updated_at` field to ensure consistency between CLI edits and manual file edits.

# Problem Statement

Currently, `updated_at` timestamps behave inconsistently:
- **CLI edits (`pm move`, `pm edit`)**: Update `updated_at` in YAML front-matter
- **Manual edits (text editor)**: `updated_at` remains stale until next CLI operation
- **Git clone**: Sets filesystem mtime to clone time, making all cached timestamps incorrect

This creates confusion about when tickets were actually last modified.

# Solution Approach

Treat `updated_at` as a **computed persistent cache** using git commit history with mtime fallback:

1. **Primary source (committed files):** Parse `git log --format=%ct` for last commit timestamp
2. **Fallback (uncommitted changes):** Use filesystem mtime if working tree modified
3. **Incremental strategy:** Only recompute timestamps for files that changed since last cache sync
4. **Cache storage:** Store computed timestamps in SQLite cache (already has `updated_at` column)

**Key decision:** We KEEP the `updated_at` field in YAML front-matter (don't remove it). This maintains backward compatibility and human readability. The CLI just stops writing to it, treating it as read-only legacy data.

# Performance Analysis

**Assumptions:**
- `git log` takes ~5ms per file call
- SQLite query takes <1ms
- Filesystem mtime read: <0.1ms

**Scenarios:**

| Ticket Count | Changed Files | Strategy | Time (Normal) | Time (Fresh Clone) |
|--------------|---------------|----------|---------------|---------------------|
| 100 tickets  | 5 changed     | Incremental | 25ms (5×5ms git) | 500ms (100×5ms) |
| 500 tickets  | 10 changed    | Incremental | 50ms (10×5ms git) | 2.5s (500×5ms) |
| 1000 tickets | 20 changed    | Incremental | 100ms (20×5ms git) | 5s (1000×5ms) |

**Normal operation (incremental):** Only recompute changed files → 0-100ms overhead
**Fresh clone (all stale):** Recompute all 1000 tickets → ~5 seconds one-time cost

**Conclusion:** Incremental strategy makes this O(changed files), not O(total tickets).

# Future Optimizations (Out of Scope)

This ticket implements the basic incremental strategy. Future enhancements could include:

1. **Skip "final" states:**  
   - Don't recompute timestamps for tickets in terminal states (e.g., "done", "archived")
   - Challenge: States are user-defined in `workflow.yaml` - would need `terminal: true` metadata
   - Benefit: Reduce git calls by 50%+ in mature projects

2. **Batch git query:**  
   - Replace per-file `git log` calls with single `git log --name-status --format="%H %ct"` 
   - Parse output to build timestamp map in one subprocess call
   - Benefit: Reduce 1000 git calls to 1 call (~5s → ~50ms)

3. **Commit hash cache:**  
   - Store last processed commit SHA in cache metadata
   - Skip git queries entirely if HEAD hasn't moved and file mtime unchanged
   - Benefit: Zero git overhead for unchanged repositories

4. **Configurable skip patterns:**  
   - Add `.pm/config/cache.yaml` with `skip_timestamp_update: [done, archived]`
   - Let users opt-in to skipping terminal states
   - Benefit: Performance tuning per-project

**Recommendation:** Implement basic incremental strategy now (this ticket). Defer optimizations until users report performance issues with 1000+ ticket repositories.

# Implementation Steps

- [ ] Add `GetLastModifiedTime(ticketPath string) (time.Time, error)` to `internal/ticket/`:
  - [ ] Run `git log -1 --format=%ct -- <file>` to get commit timestamp
  - [ ] If file is uncommitted (exit code 128 or no output), fall back to `os.Stat()` mtime
  - [ ] Parse Unix timestamp and return `time.Time`
- [ ] Update `internal/cache/sync.go`:
  - [ ] In `syncTicketToCache()`, replace `ticket.UpdatedAt` with `GetLastModifiedTime(path)`
  - [ ] Store computed timestamp in cache's `updated_at` column
  - [ ] Add debug logging: `"Computed updated_at for TICKET-123: 2026-02-04 from git"`
- [ ] Update CLI commands to stop writing `updated_at`:
  - [ ] `cmd/pm/move.go`: Remove `ticket.UpdatedAt = time.Now()` line
  - [ ] `cmd/pm/edit.go`: Remove `updated_at` update logic
  - [ ] Keep `created_at` auto-update (still needed for new tickets)
- [ ] Update `pm show` to display computed timestamp:
  - [ ] Read `updated_at` from cache (already does this)
  - [ ] No code change needed - cache already has computed value

# Unit Tests

- [ ] Test `GetLastModifiedTime()`:
  - [ ] Committed file: Returns git commit timestamp
  - [ ] Modified file: Returns filesystem mtime (newer than commit)
  - [ ] Untracked file: Returns mtime
  - [ ] Non-existent file: Returns error
- [ ] Test git command parsing:
  - [ ] Valid Unix timestamp: "1706853600" → 2024-02-02T00:00:00Z
  - [ ] Empty output: Falls back to mtime
  - [ ] Git not available: Falls back to mtime (graceful degradation)

# Integration Tests

Add to `integration_test.go`:

- [ ] Create ticket, commit it, verify `pm show` displays git commit time
- [ ] Edit ticket file manually (no git commit), verify `pm show` displays mtime
- [ ] Move ticket with `pm move`, verify timestamp updates to latest commit
- [ ] Modify ticket and commit with custom timestamp (e.g., 1 day ago), verify shows correct time

# Acceptance Criteria

- [ ] `pm show TICKET-123` displays timestamp from git commit (if committed)
- [ ] `pm show TICKET-123` displays mtime (if modified but not committed)
- [ ] Manual file edits are correctly reflected in `pm list` and `pm show`
- [ ] Fresh git clone shows correct historical timestamps (not clone time)
- [ ] Performance: <100ms overhead for `pm list` with 20 changed tickets
- [ ] CLI commands (`pm move`, `pm edit`) no longer write to `updated_at` field in YAML
- [ ] Existing tickets with `updated_at` in YAML are not broken (backward compatible)
- [ ] All unit tests pass
- [ ] All integration tests pass

# Notes

- **Backward compatibility:** Old tickets with `updated_at` in YAML won't break - we just ignore that field and use git/mtime
- **Migration path:** No migration needed - system automatically computes correct timestamps on next cache sync
- **Git dependency:** This makes git a hard requirement (already is for the project, so acceptable)
- **Future-proof:** Incremental strategy keeps performance acceptable even at 10,000+ tickets

