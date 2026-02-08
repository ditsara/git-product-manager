---
id: GPM-2
title: "Stage 3: Advanced Relationships and Search"
type: epic
status: backlog
priority: medium
points: 0
parent: ""
depends_on: [GPM-1]
blocks: []
related: []
labels: [relationships, search, visualization]
assignee: ""
created_at: 2026-02-02T03:57:21Z
updated_at: 2026-02-02T03:57:21Z
---

# Description

[Sonnet 4.5]

This final stage completes the vision by adding powerful relationship tracking, visualization, and efficient searching capabilities.

## Objectives

- Enable complex ticket relationships (hierarchies, dependencies, blocking)
- Provide visual representation of ticket structures
- Implement fast full-text search across tickets
- Add relationship validation and integrity checks

## Child Tickets

- **GPM-45**: Implement `pm link` and `pm unlink` with automatic symmetry
- **GPM-46**: Cache sync: detect and warn on broken relationship symmetry
- **GPM-47**: Implement `pm blocked` for dependency tracking
- **GPM-48**: Implement `pm tree` for hierarchy visualization
- **GPM-49**: Implement `pm search` for full-text search
- **GPM-50**: Enhance `pm list` with relationship-aware filtering

Create and manage array-based relationships between tickets with automatic bidirectional symmetry for dependency pairs.

**Note on `parent` field:**
- The `parent` field is excluded from link/unlink (single-value, not array)
- Manage via: `pm edit <id> --field parent=<parent-id>`

### 2. Hierarchy Visualization (`pm tree`)
Display ticket hierarchies as ASCII trees showing parent-child relationships recursively.

### 3. Dependency Tracking (`pm blocked`)
**See: GPM-47**

Show blocking relationships and identify bottlenecks in the workflow.

### 4. Full-Text Search (`pm search`)
Implement efficient search across tickets using SQLite FTS5.

### 5. Enhanced List Filtering
**Note**: `pm list --parent <id>` is already implemented in Stage 1.

Enhance with additional relationship-aware filtering.

## Validation Enhancements

### Reference Integrity
- **All referenced IDs must exist**: depends_on, blocks, related fields (parent validated separately via pm edit)
- **No self-reference**: Ticket cannot reference itself in any relationship field
- **Circular dependency detection**: Prevent cycles in depends_on relationships
- **Array operations only**: link/unlink commands only operate on array fields (depends_on, blocks, related)

### User Responsibility
- **No semantic enforcement**: System doesn't enforce logical hierarchies (e.g., stories containing epics)
- **User maintains**: Teams responsible for logical relationship structures
- **Parent management**: Users set/clear parent field explicitly via `pm edit --field parent=...`

## Testing Requirements

### Unit Tests
- [ ] Relationship validation (circular dependencies, self-reference)
- [ ] Ticket reference integrity checking
- [ ] Tree rendering algorithm (various depths and structures)
- [ ] FTS query parsing and execution
- [ ] Parent filtering logic

### Integration Tests
- [ ] Link and unlink tickets (all relationship types)
- [ ] Visualize ticket hierarchy with `pm tree`
- [ ] Query blocked tickets and dependencies
- [ ] Full-text search across ticket content
- [ ] Filter tickets by parent with `pm list --parent`
- [ ] Combined filters (parent + status + label)

## Acceptance Criteria

- [ ] Users can create and remove all relationship types
- [ ] Ticket hierarchies display correctly as ASCII trees
- [ ] Dependency information is easily accessible
- [ ] Search finds tickets by content quickly
- [ ] Parent-based filtering works correctly
- [ ] Circular dependencies are detected and prevented
- [ ] All relationship references are validated
- [ ] All tests pass (unit + integration)
- [ ] Documentation updated with new commands

## Implementation Checklist

- [ ] **`pm link` & `pm unlink`**: Implement array relationship types (`depends-on`, `blocks`, `related`) - parent excluded by design
- [ ] **`pm tree`**: Visualize the parent-child hierarchy
- [ ] **`pm blocked`**: Show dependency information
- [ ] **`pm search`**: Implement full-text search using SQLite FTS table
- [ ] **`pm list`**: Update to add `--parent` filtering
- [ ] **Validation**: Add reference integrity and circular dependency checks
- [ ] **Database**: Add migration for `relationships` and `tickets_fts` tables
- [ ] **Database**: Update cache logic to index all relationships
- [ ] **Tests**: Unit tests for relationship validation (circular dependencies, self-reference)
- [ ] **Tests**: Integration tests for linking, unlinking, and visualizing tickets (`tree`, `blocked`)
- [ ] **Tests**: Tests for `pm search` and advanced filtering
