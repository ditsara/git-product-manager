---
assignee: ""
blocks: []
created_at: "2026-02-14T08:36:12Z"
depends_on:
    - GPM-62
id: GPM-63
labels:
    - database
    - refactoring
parent: GPM-61
points: 5
priority: medium
related: []
status: done
title: Refactor internal/cache/sync.go to use Bob
type: task
updated_at: "2026-02-14T10:46:42Z"
---


# Description

**[Claude Sonnet 4.5]**

## Problem Statement

Refactor `internal/cache/sync.go` to use Bob ORM for bulk insert operations. This file contains the most repetitive SQL code with prepared statements for tickets, comments, and relationships.

## Current State

**File:** `internal/cache/sync.go`  
**Lines:** ~100-270

**SQL operations:**
- DELETE FROM tickets/comments/relationships
- Prepared INSERT statements (3 different tables)
- Bulk inserts in transaction
- UPDATE cache_metadata

**Complexity:** Medium - bulk operations with prepared statements

## Solution Approach

Replace raw SQL prepared statements with Bob's bulk insert capabilities.

## Edge Cases

- Ensure bulk inserts are as efficient as prepared statements
- Maintain transaction atomicity (BEGIN/COMMIT/ROLLBACK)
- Handle empty arrays gracefully (no tickets/comments/relationships)
- Preserve error messages for debugging

## Implementation Steps

- [x] Refactor DELETE operations to use Bob
- [x] Refactor ticket INSERT to use Bob bulk insert
- [x] Refactor comment INSERT to use Bob bulk insert
- [x] Refactor relationship INSERT to use Bob bulk insert
- [x] Refactor cache_metadata UPDATE
- [x] Run `internal/cache/sync_test.go`
- [x] Run integration tests that trigger cache sync
- [x] Verify performance is comparable

## Acceptance Criteria

- [x] All SQL in `sync.go` migrated to Bob
- [x] `internal/cache/sync_test.go` passes unchanged
- [x] Integration tests pass
- [x] Transaction semantics preserved
- [x] Code is significantly cleaner
- [x] No performance regression

## Implementation Results

**Completed:** 2026-02-14

**Changes Made:**
1. **Imports:** Added `context`, `bob/dialect/sqlite`, `dm`, `im`
2. **DELETE operations:** Converted to `sqlite.Delete(dm.From(table))`
3. **INSERT operations:** Restructured to collect data first, then bulk insert
4. **UPDATE operation:** Converted to `sqlite.Insert(..., im.OrReplace())`

**Code Structure Change:**
- **Before:** Prepared statements with immediate execution in loops
- **After:** Collect all data in slices, then single bulk INSERT per table

**Benefits Achieved:**
- ✅ **Cleaner code:** No manual SQL string construction
- ✅ **Type safety:** Programmatic query building
- ✅ **Bulk operations:** More efficient than prepared statement loops
- ✅ **Maintainability:** Easier to understand data collection vs execution phases
- ✅ **Transaction safety:** All Bob queries execute within existing transaction

**Test Results:**
- ✅ All cache tests pass (8/8 tests, 1.93s)
- ✅ Integration tests pass (blocked, assign, show, etc.)
- ✅ No behavioral changes
- ✅ No performance regression (similar execution time)

**Pattern Established:**
1. Use `dm.*` for DELETE queries
2. Use `im.*` for INSERT queries
3. For bulk inserts: collect data first, then build single query with multiple `.Apply(im.Values(...))`
4. Build queries with `query.Build(ctx)`, execute with `tx.Exec(sql, args...)`
