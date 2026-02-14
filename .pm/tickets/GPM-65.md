---
id: GPM-65
title: "Refactor cmd/pm/list.go to use Bob"
type: task
status: backlog
priority: medium
points: 8

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-61"
depends_on: [GPM-64]
blocks: []
related: []

labels: [database, refactoring]
assignee: ""
created_at: "2026-02-14T08:39:52Z"
updated_at: "2026-02-14T08:39:52Z"
---

# Description

**[Claude Sonnet 4.5]**

## Problem Statement

Refactor `cmd/pm/list.go` to use Bob ORM for the most complex SQL in the codebase: recursive CTEs with dynamic query building. This is the final and most challenging migration.

## Current State

**File:** `cmd/pm/list.go`  
**Lines:** ~119-191 (query building)

**SQL operations:**
- Recursive CTEs (WITH RECURSIVE subtree)
- Dynamic query construction based on flags
- Multiple query paths (parent filter, all, default)
- Dynamic WHERE clause building
- Nested subqueries (has_children check)

**Complexity:** Very High - dynamic recursive CTEs

## Solution Approach

This is the most complex refactor due to:
1. **Recursive CTE** - `WITH RECURSIVE subtree AS (...)`
2. **Dynamic query building** - Different queries based on flags
3. **Programmatic filters** - Status/label filters added conditionally

**Strategy:**
- Use Bob's CTE builder for recursive queries
- Build query programmatically based on flags
- Demonstrate Bob's dynamic query capabilities

## Edge Cases

- Recursive CTE must terminate correctly
- Empty result sets handled gracefully
- Case-insensitive parent matching (UPPER())
- Dynamic status filters (includeStates vs excludeStates)
- Subquery for has_children indicator
- Default behavior (top-level tickets only)

## Implementation Steps

- [ ] Study Bob's CTE/WITH support
- [ ] Refactor base query structure
  - Handle --parent with/without --all flags
  - Default top-level query
- [ ] Implement recursive CTE using Bob
  - WITH RECURSIVE subtree logic
  - Recursive union
- [ ] Add dynamic status filtering
  - includeStates array
  - excludeStates array
  - Programmatic WHERE clause building
- [ ] Add has_children subquery
  - Nested EXISTS check
  - Case-insensitive matching
- [ ] Run `integration_list_test.go`
- [ ] Manually test all flag combinations
- [ ] Verify performance (CTEs can be slow)

## Acceptance Criteria

- [ ] All SQL in `list.go` migrated to Bob
- [ ] `integration_list_test.go` passes unchanged
- [ ] Recursive CTE works correctly
- [ ] All flag combinations produce correct results
- [ ] Dynamic query building is cleaner than string concatenation
- [ ] Code readability improved
- [ ] No performance regression

## Example Challenge

**Current approach (lines 123-134):**
```go
query = `
    WITH RECURSIVE subtree(id) AS (
        SELECT id FROM tickets WHERE UPPER(parent) = UPPER(?)
        UNION ALL
        SELECT t.id FROM tickets t
        JOIN subtree s ON UPPER(t.parent) = UPPER(s.id)
    )
    SELECT id, title, type, status,
        CASE WHEN EXISTS(SELECT 1 FROM tickets AS t WHERE UPPER(t.parent) = UPPER(tickets.id))
            THEN 1 ELSE 0 END AS has_children
    FROM tickets
    WHERE id IN (SELECT id FROM subtree)`
```

**Need Bob equivalent** - demonstrate Bob can handle this!

## Notes

- This is the ultimate test of Bob's capabilities
- Most challenging migration due to complexity
- Success here proves Bob can handle anything in this codebase
- May discover Bob limitations - document them
- Consider performance implications of CTEs
