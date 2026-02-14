---
id: GPM-62
title: "Initial Bob integration + POC refactor"
type: task
status: backlog
priority: high
points: 3

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-61"
depends_on: []
blocks: [GPM-63, GPM-64, GPM-65]
related: []

labels: [database, refactoring]
assignee: ""
created_at: "2026-02-14T08:35:06Z"
updated_at: "2026-02-14T08:35:06Z"
---

# Description

**[Claude Sonnet 4.5]**

## Problem Statement

Setup Bob ORM infrastructure and prove it works with a simple proof-of-concept refactor. This establishes the foundation for all subsequent migration work.

## Solution Approach

**Phase 1: Setup Bob**
1. Add Bob dependency to `go.mod`
2. Create basic Bob configuration (if needed)
3. Setup SQLite dialect

**Phase 2: POC Refactor**
Refactor the simplest SQL query as proof-of-concept:
- **File:** `cmd/pm/blocked.go`
- **Line:** 185
- **Query:** `SELECT title, status FROM tickets WHERE id = ?`

This is a perfect POC because:
- Simple SELECT with WHERE clause
- Single table, no joins
- Existing integration tests cover it (`integration_blocked_test.go`)
- Low risk, easy to verify

**Phase 3: Verify**
- Run integration tests
- Ensure behavior is identical
- Document any learnings

## Edge Cases

- Ensure Bob's SQLite dialect handles parameterized queries correctly
- Verify error handling (sql.ErrNoRows) works the same way
- Confirm Bob works with existing `database/sql` connections (mixed usage during migration)

## Implementation Steps

- [ ] Add Bob to dependencies
  ```bash
  go get github.com/stephenafamo/bob
  ```
- [ ] Import Bob SQLite packages
  ```go
  import (
      "github.com/stephenafamo/bob/dialect/sqlite"
      "github.com/stephenafamo/bob/dialect/sqlite/sm"
  )
  ```
- [ ] Refactor ticket lookup query in `blocked.go` line 185
  - Replace raw SQL with Bob query builder
  - Keep same error handling logic
  - Maintain same return values
- [ ] Run `integration_blocked_test.go`
- [ ] Verify no behavior changes
- [ ] Document Bob usage pattern for team

## Acceptance Criteria

- [ ] Bob added to `go.mod` and imports successfully
- [ ] Simple SELECT query refactored to use Bob Layer 1 (query builder)
- [ ] `integration_blocked_test.go` passes unchanged
- [ ] Code is cleaner/more readable than raw SQL
- [ ] Pattern is documented for future refactors

## Example (POC Implementation)

**Before:**
```go
var title, status string
err := db.QueryRow("SELECT title, status FROM tickets WHERE id = ?", ticketID).Scan(&title, &status)
if err == sql.ErrNoRows {
    log.Fatalf("Ticket not found: %s", ticketID)
}
```

**After (Bob Layer 1):**
```go
query := sqlite.Select(
    sm.Columns("title", "status"),
    sm.From("tickets"),
    sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(ticketID))),
)

var title, status string
err := query.Query(ctx, db).Scan(&title, &status)
if err == sql.ErrNoRows {
    log.Fatalf("Ticket not found: %s", ticketID)
}
```

**Benefits demonstrated:**
- ✅ More readable (no SQL string)
- ✅ Type-safe column/table references
- ✅ Programmatic query building
- ⚠️ Slightly more verbose (acceptable trade-off)

## Notes

This ticket is the **proof of concept** for the entire migration. If this doesn't work smoothly, we reassess the strategy before proceeding to GPM-63+.