---
id: GPM-63
title: "Refactor internal/cache/sync.go to use Bob"
type: task
status: backlog
priority: medium
points: 5

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-61"
depends_on: [GPM-62]
blocks: []
related: []

labels: [database, refactoring]
assignee: ""
created_at: "2026-02-14T08:36:12Z"
updated_at: "2026-02-14T08:36:12Z"
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

- [ ] Refactor DELETE operations to use Bob
- [ ] Refactor ticket INSERT to use Bob bulk insert
- [ ] Refactor comment INSERT to use Bob bulk insert
- [ ] Refactor relationship INSERT to use Bob bulk insert
- [ ] Refactor cache_metadata UPDATE
- [ ] Run `internal/cache/sync_test.go`
- [ ] Run integration tests that trigger cache sync
- [ ] Verify performance is comparable

## Acceptance Criteria

- [ ] All SQL in `sync.go` migrated to Bob
- [ ] `internal/cache/sync_test.go` passes unchanged
- [ ] Integration tests pass
- [ ] Transaction semantics preserved
- [ ] Code is significantly cleaner
- [ ] No performance regression
