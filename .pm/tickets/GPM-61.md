---
assignee: ""
blocks: []
created_at: "2026-02-14T08:34:21Z"
depends_on: []
id: GPM-61
labels:
    - database
    - refactoring
parent: ""
points: 13
priority: high
related:
    - GPM-60
status: done
title: Migrate to Bob ORM
type: epic
updated_at: "2026-02-14T11:00:48Z"
---


# Description

**[Claude Sonnet 4.5]**

Incrementally migrate from raw `database/sql` to Bob ORM for improved type safety, reduced boilerplate, and better testability.

## Background

Following the ORM evaluation in GPM-60, we've chosen **Bob (stephenafamo/bob)** for progressive adoption:
- ✅ Perfect golang-migrate compatibility
- ✅ Handles all edge cases (CTEs, joins, bulk inserts, dynamic queries)
- ✅ Progressive adoption (Layer 1 → Layer 2 as needed)
- ✅ Factory generation for testing

## Strategy

**Progressive refactoring** from simplest to most complex:
1. **Initial integration** - Add Bob, configure, POC with simplest query
2. **Bulk operations** - Refactor cache sync (prepared statements)
3. **Complex joins** - Refactor blocked command (aggregations)
4. **Dynamic CTEs** - Refactor list command (recursive CTEs, filters)

**Integration test strategy:**
- ✅ **Ideally:** Integration tests remain functionally unchanged
- ⚠️ **May modify if:** Bob requires different assertion patterns or setup
- 🎯 **Goal:** Ensure behavior is identical before/after each refactor

## Implementation Philosophy

**Bob Layers (use appropriately):**
- **Layer 1 (Query Builder):** For most queries - programmatic, readable, no code generation
- **Layer 2 (Generated Models):** Optional - full type safety when schema is stable
- **Layer 3 (Factories):** For test data generation
- **Layer 4 (Query Gen):** For hand-crafted SQL (if needed)

**Start with Layer 1 only** - add Layer 2 (code generation) later if beneficial.

## Success Criteria

- [x] All SQL queries evaluated for Bob migration
- [x] Simple CRUD operations migrated to Bob where beneficial
- [x] Complex/dynamic queries kept as raw SQL with justification
- [x] Integration tests pass without functional changes
- [x] No performance regression
- [x] Code maintainability improved where Bob was used

## Child Tickets

1. **GPM-62:** ✅ Initial Bob integration + POC (simple SELECT) - **DONE**
2. **GPM-63:** ✅ Refactor cache sync (bulk inserts/deletes) - **DONE**
3. **GPM-64:** ⚠️ Refactor blocked command - **PARTIAL** (simple query only, see GPM-67)
4. **GPM-65:** ❌ Refactor list command - **NOT SUITABLE** (kept as raw SQL)

## Notes

- Keep golang-migrate unchanged (Bob doesn't manage migrations)
- Each ticket should be independently testable
- Can pause/reassess after POC if issues arise
- Document any deviations from original test strategy

## Implementation Results

**Completed:** 2026-02-14

**Bob Successfully Used For:**
- ✅ Simple SELECT queries (GPM-62: blocked.go line 185)
- ✅ Bulk DELETE operations (GPM-63: cache sync cleanup)
- ✅ Bulk INSERT operations (GPM-63: cache population)

**Bob NOT Suitable For:**
- ❌ SQLite-specific aggregations (GROUP_CONCAT)
- ❌ Complex JOIN queries with dynamic HAVING clauses (GPM-67 created for future investigation)
- ❌ Recursive CTEs with dynamic query paths (list.go)
- ❌ Dynamic query building with string manipulation (WHERE clause assembly)

**Key Learnings:**

1. **Bob excels at straightforward CRUD:**
   - Cache sync operations were much cleaner with Bob
   - Reduced boilerplate for bulk inserts/deletes
   - Better readability than raw SQL for simple queries

2. **Raw SQL is better for complex/dynamic queries:**
   - Recursive CTEs are SQLite-specific
   - Dynamic query path selection (4 different queries in list.go)
   - Programmatic WHERE clause building
   - Using Bob's Raw() defeats the purpose of type safety

3. **Decision Framework:**
   - Use Bob: Simple CRUD, bulk operations, static queries
   - Use raw SQL: Complex queries, SQLite features, dynamic query paths
   - Avoid: Bob's Raw() - if you need it, just use raw SQL

**Test Results:**
- ✅ All integration tests pass
- ✅ No performance regression
- ✅ Code quality improved where Bob was used

**Recommendation:**
Continue using Bob for new CRUD operations in cache layer and simple queries. Keep complex/dynamic queries as raw SQL. Don't force Bob where it doesn't add value.

