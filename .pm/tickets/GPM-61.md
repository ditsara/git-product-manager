---
id: GPM-61
title: "Migrate to Bob ORM"
type: epic
status: backlog
priority: high
points: 13

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""
depends_on: []
blocks: []
related: [GPM-60]

labels: [database, refactoring]
assignee: ""
created_at: "2026-02-14T08:34:21Z"
updated_at: "2026-02-14T08:34:21Z"
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

- [x] All SQL queries migrated to Bob (or consciously kept as raw SQL)
- [x] Integration tests pass without functional changes
- [x] No performance regression
- [x] Code is more maintainable (less boilerplate)
- [x] Dynamic query building is cleaner (list.go filters)

## Child Tickets

1. **GPM-62:** Initial Bob integration + POC (simplest query)
2. **GPM-63:** Refactor cache sync (bulk inserts)
3. **GPM-64:** Refactor blocked command (complex joins)
4. **GPM-65:** Refactor list command (CTEs, dynamic filters)

## Notes

- Keep golang-migrate unchanged (Bob doesn't manage migrations)
- Each ticket should be independently testable
- Can pause/reassess after POC if issues arise
- Document any deviations from original test strategy

