---
assignee: ""
blocks:
    - GPM-63
    - GPM-64
    - GPM-65
created_at: "2026-02-14T08:35:06Z"
depends_on: []
id: GPM-62
labels:
    - database
    - refactoring
parent: GPM-61
points: 3
priority: high
related: []
status: done
title: Initial Bob integration + POC refactor
type: task
updated_at: "2026-02-14T10:33:35Z"
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

- [x] Add Bob to dependencies
  ```bash
  go get github.com/stephenafamo/bob
  ```
- [x] Import Bob SQLite packages
  ```go
  import (
      "context"
      "github.com/stephenafamo/bob/dialect/sqlite"
      "github.com/stephenafamo/bob/dialect/sqlite/sm"
  )
  ```
- [x] Refactor ticket lookup query in `blocked.go` line 185
  - Replace raw SQL with Bob query builder
  - Keep same error handling logic
  - Maintain same return values
- [x] Run `integration_blocked_test.go`
- [x] Verify no behavior changes
- [x] Document Bob usage pattern for team

## Acceptance Criteria

- [x] Bob added to `go.mod` and imports successfully
- [x] Simple SELECT query refactored to use Bob Layer 1 (query builder)
- [x] `integration_blocked_test.go` passes unchanged
- [x] Code is cleaner/more readable than raw SQL
- [x] Pattern is documented for future refactors

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

## Implementation Results

**Completed:** 2026-02-14

**Final Implementation:**
```go
// Verify ticket exists using Bob query builder
query := sqlite.Select(
    sm.Columns("title", "status"),
    sm.From("tickets"),
    sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(ticketID))),
)

// Build the SQL query
querySQL, args, err := query.Build(context.Background())
if err != nil {
    log.Fatalf("Error building query: %v", err)
}

var title, status string
err = db.QueryRow(querySQL, args...).Scan(&title, &status)
if err == sql.ErrNoRows {
    log.Fatalf("Ticket not found: %s", ticketID)
} else if err != nil {
    log.Fatalf("Error querying ticket: %v", err)
}
```

**Key Learnings:**
1. ✅ **Bob API:** Use `query.Build(ctx)` to generate SQL, then execute with `db.QueryRow()`
2. ✅ **Context required:** Bob requires `context.Context` for all query building
3. ✅ **Mixed usage works:** Bob and raw SQL can coexist during incremental migration
4. ✅ **Error handling unchanged:** `sql.ErrNoRows` works identically
5. ✅ **Zero test changes:** Integration test passed without modification

**Pattern for Future Refactors:**
1. Build query using Bob dialect (e.g., `sqlite.Select()`)
2. Apply mods (columns, from, where, etc.) using `sm.*` functions
3. Call `query.Build(context.Background())` to get SQL + args
4. Execute with standard `database/sql` methods (`QueryRow`, `Query`, `Exec`)
5. Error handling remains identical to raw SQL

**Dependencies Added:**
- `github.com/stephenafamo/bob v0.42.0`
- `github.com/stephenafamo/bob/orm v0.42.0` (transitive, required for build)
- `github.com/aarondl/opt v0.0.0-20250607033636-982744e1bd65` (transitive)