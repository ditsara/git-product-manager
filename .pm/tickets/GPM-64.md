---
id: GPM-64
title: "Refactor cmd/pm/blocked.go to use Bob"
type: task
status: backlog
priority: medium
points: 5

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-61"
depends_on: [GPM-63]
blocks: []
related: []

labels: [database, refactoring]
assignee: ""
created_at: "2026-02-14T08:38:45Z"
updated_at: "2026-02-14T08:38:45Z"
---

# Description

**[Claude Sonnet 4.5]**

## Problem Statement

Refactor `cmd/pm/blocked.go` to use Bob ORM for complex JOIN queries with GROUP_CONCAT aggregations. This demonstrates Bob's capability with advanced SQL features.

## Current State

**File:** `cmd/pm/blocked.go`  
**Functions:** `showGlobalBlockedView()` (line 79), `showTicketBlockedView()` (line 182)

**SQL operations:**
- Complex multi-table JOINs
- GROUP BY with GROUP_CONCAT aggregations
- Subqueries with HAVING clauses
- Multiple related queries (dependencies, blockers)

**Complexity:** High - complex joins and aggregations

## Solution Approach

Refactor both functions to use Bob's query builder while maintaining readability:
- Use Bob's join methods
- Preserve aggregation logic
- Keep the same result structure

**Note:** Line 185 simple SELECT was already refactored in GPM-62 (POC).

## Edge Cases

- Ensure GROUP_CONCAT syntax works correctly
- Handle empty result sets (no blocked tickets)
- Preserve NULL handling in joins
- Maintain case-insensitive matching (UPPER())
- Complex string parsing (splitting GROUP_CONCAT results)

## Implementation Steps

- [ ] Refactor `showGlobalBlockedView()` complex query
  - Multi-table JOIN with GROUP_CONCAT
  - HAVING clause with aggregation
- [ ] Refactor `showTicketBlockedView()` dependency queries
  - Two separate JOIN queries (depends_on, blocks)
- [ ] Verify all edge cases handled
- [ ] Run `integration_blocked_test.go`
- [ ] Manually test various blocked scenarios

## Acceptance Criteria

- [ ] All SQL in `blocked.go` migrated to Bob
- [ ] `integration_blocked_test.go` passes unchanged
- [ ] Complex joins work correctly
- [ ] GROUP_CONCAT aggregations produce same results
- [ ] Code readability maintained or improved
- [ ] No performance regression

## Notes

- This is the most complex query migration
- Demonstrates Bob can handle real-world SQL complexity
- May need to fallback to raw SQL for parts if Bob can't express it
- Document any Bob limitations discovered
