---
id: GPM-67
title: "Migrate blocked.go JOIN queries to Bob"
type: task
status: backlog
priority: low
points: 2

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-61"
depends_on: []
blocks: []
related: [GPM-64]

labels: [database, refactoring, technical-debt]
assignee: ""
created_at: "2026-02-14T10:57:04Z"
updated_at: "2026-02-14T10:57:04Z"
---

# Description

**[Claude Sonnet 4.5]**

## Problem Statement

During GPM-64, we discovered that Bob's Raw() function defeats the purpose of using an ORM. The simple JOIN queries in `showTicketBlockedView()` were kept as raw SQL instead of being migrated to Bob.

This ticket tracks the proper migration of these queries once we understand Bob's API better.

## Current State

**File:** `cmd/pm/blocked.go`
**Functions:** `showTicketBlockedView()` - two JOIN queries (depends_on, blocks)

**SQL operations:**
- Simple 2-table JOIN with WHERE clause
- Basic parameterization
- ORDER BY

## Solution Approach

Research Bob's proper API for:
1. Table aliases without quotes
2. JOIN conditions without Raw()
3. WHERE clauses with AND without Raw()
4. Proper parameter binding

The goal is to use Bob's type-safe query builder without falling back to Raw().

## Acceptance Criteria

- [ ] Both JOIN queries in `showTicketBlockedView()` use Bob without Raw()
- [ ] integration_blocked_test.go passes unchanged
- [ ] No quoted identifiers in generated SQL (avoid case-sensitivity issues)
- [ ] Document the proper Bob pattern for simple JOINs

## Notes

- Low priority - raw SQL works fine for now
- This is a learning opportunity for Bob's API
- The initial attempt in GPM-64 used Raw() which was rejected
- May require Bob documentation review or examples