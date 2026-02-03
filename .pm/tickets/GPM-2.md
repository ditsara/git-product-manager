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

## Key Features

### 1. Relationship Management (`pm link` & `pm unlink`)
Create and manage relationships between tickets:

**`pm link <id> <target-id> [--type TYPE]`**:
- **Types**: `parent`, `depends-on`, `blocks`, `related`
- **Default**: `related`
- **Behavior**: Updates the appropriate field in the source ticket
- **Auto-commit**: `chore(pm): Link {id} to {target-id} ({type})`

**`pm unlink <id> <target-id> [--type TYPE]`**:
- **Behavior**: Removes the target from the relationship array
- **No type specified**: Removes from all relationship fields

### 2. Hierarchy Visualization (`pm tree`)
Display ticket hierarchies as ASCII trees:

**`pm tree <id> [--depth N]`**:
- **Display**: Shows parent-child relationships recursively
- **Default depth**: Unlimited
- **Example output**:
  ```
  EPIC-123: Implement Authentication
  ├── STORY-456: OAuth2 Login
  │   ├── TASK-789: Setup Google Provider
  │   └── TASK-790: Create JWT Middleware
  └── STORY-457: Password Reset Flow
      └── TASK-791: Email Template
  ```

### 3. Dependency Tracking (`pm blocked`)
Show blocking relationships:

**`pm blocked [<id>]`**:
- **No ID**: List all tickets with unresolved dependencies
- **With ID**: Show what blocks this ticket and what it blocks
- **Use case**: Identify bottlenecks in workflow

### 4. Full-Text Search (`pm search`)
Implement efficient search across tickets:

**`pm search <query>`**:
- **Backend**: Use SQLite FTS5 (Full-Text Search) on title and body
- **Scope**: Search across all ticket content
- **Performance**: Fast indexed queries
- **Results**: Display matching tickets with context

### 5. Enhanced List Filtering (`pm list --parent`)
Add parent-based filtering:

**`pm list --parent <id>`**:
- **Display**: Show all tickets with the specified parent
- **Use case**: View all stories/tasks under an epic
- **Combination**: Works with existing filters (status, label, assignee)

## Database Changes

### New Migration (000004)
- **Add `relationships` table**:
  ```sql
  CREATE TABLE relationships (
    from_ticket TEXT NOT NULL,
    to_ticket TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    PRIMARY KEY (from_ticket, to_ticket, relationship_type)
  );
  CREATE INDEX idx_from ON relationships(from_ticket);
  CREATE INDEX idx_to ON relationships(to_ticket);
  ```

- **Add `tickets_fts` table** for full-text search:
  ```sql
  CREATE VIRTUAL TABLE tickets_fts USING fts5(
    id, title, body,
    content='tickets'
  );
  ```

### Cache Logic Updates
- **Index relationships**: Populate relationships table when syncing
- **Update FTS**: Keep full-text search index synchronized
- **Query optimization**: Use indexes for fast relationship queries

## Validation Enhancements

### Reference Integrity
- **All referenced IDs must exist**: parent, depends_on, blocks, related fields
- **No self-reference**: Ticket cannot reference itself
- **Circular dependency detection**: Prevent cycles in depends_on relationships

### User Responsibility
- **No semantic enforcement**: System doesn't enforce logical hierarchies (e.g., stories containing epics)
- **User maintains**: Teams responsible for logical relationship structures

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

- [ ] **`pm link` & `pm unlink`**: Implement all relationship types (`parent`, `depends-on`, `blocks`, `related`)
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
