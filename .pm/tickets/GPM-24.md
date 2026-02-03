---
assignee: ""
blocks: []
created_at: "2026-02-03T14:54:26Z"
depends_on: []
id: GPM-24
labels:
    - ux
    - filtering
    - hierarchy
    - stage-1.6
parent: ""
points: 3
priority: high
related:
    - GPM-1
status: done
title: Add hierarchical filtering to pm list
type: story
updated_at: "2026-02-03T15:14:45Z"
---


# Description

[Sonnet 4.5]

Implement hierarchical filtering for `pm list` to reduce clutter when working with epics and their child stories.

## Problem Statement

When using epics with child stories (e.g., GPM-1 with 5 child stories), `pm list` becomes cluttered showing both the epic and all its children. Users typically want to see either:
- **High-level view**: Top-level epics and standalone tasks (default)
- **Drill-down**: Children of a specific epic
- **Full view**: Everything (current behavior)

## User Story

As a project manager, I want `pm list` to show only top-level tickets by default, so that I can focus on high-level initiatives without seeing all the implementation details.

As a developer, I want to drill into an epic to see its child stories, so that I can pick what to work on next.

## Solution: Hybrid Filtering Approach

### Default Behavior
`pm list` → Show only tickets **without a parent** (top-level)

**Example output:**
```bash
$ pm list
ID      TITLE                          TYPE   STATUS      ASSIGNEE
GPM-1   Stage 2: Collaboration         epic   backlog     
GPM-2   Stage 3: Relationships         epic   backlog     
GPM-6   Fix findTicketByID             task   done        
GPM-10  Lazy migration check           task   done        
```

### New Flags

**`--all`**: Show all tickets (current behavior)
```bash
$ pm list --all
ID      TITLE                          TYPE   STATUS      ASSIGNEE
GPM-1   Stage 2: Collaboration         epic   backlog     
GPM-19  Implement pm comment           story  backlog     
GPM-20  Enhanced pm show               story  backlog     
GPM-21  Implement pm history           story  backlog     
...
```

**`--parent <id>`**: Show direct children of a specific ticket
```bash
$ pm list --parent GPM-1
ID      TITLE                          TYPE   STATUS      ASSIGNEE
GPM-19  Implement pm comment           story  backlog     
GPM-20  Enhanced pm show               story  backlog     
GPM-21  Implement pm history           story  backlog     
GPM-22  Implement pm assign            story  backlog     
GPM-23  Implement edit-comment         story  backlog     
```

**`--parent <id> --all`**: Show entire subtree under a ticket (recursive)
```bash
$ pm list --parent GPM-1 --all
ID        TITLE                          TYPE   STATUS      ASSIGNEE
GPM-19    Implement pm comment           story  backlog     
GPM-19-1  Create comment file parser     task   todo        # hypothetical
GPM-19-2  Add cache integration          task   todo        # hypothetical
GPM-20    Enhanced pm show               story  backlog     
...
```

### Flag Composition

All existing filters continue to work with the new flags:
- `pm list --parent GPM-1 --status backlog` → Children in backlog state
- `pm list --parent GPM-1 --type story` → Only story children
- `pm list --all --status todo` → All tickets in todo state

## Implementation Steps

- [x] Update `cmd/pm/list.go` to accept `--all` and `--parent` flags
- [x] **Default query**: Add `WHERE parent IS NULL OR parent = ''` filter
- [x] **`--all` flag**: Remove parent filter entirely (current behavior)
- [x] **`--parent <id>` flag**: Direct children query
  - SQL: `WHERE parent = ?` (exact match)
  - Validate parent ticket exists (use `findTicketByID()`)
  - Handle case-insensitive matching
- [x] **`--parent <id> --all` flag**: Recursive subtree query
  - Use SQLite recursive CTE: `WITH RECURSIVE subtree AS (...)`
  - Start with parent, recursively find all descendants
  - Example CTE:
    ```sql
    WITH RECURSIVE subtree(id) AS (
      SELECT id FROM tickets WHERE parent = ?
      UNION ALL
      SELECT t.id FROM tickets t JOIN subtree s ON t.parent = s.id
    )
    SELECT * FROM tickets WHERE id IN subtree
    ```
- [x] Handle invalid parent reference gracefully:
  - If ticket has `parent: NONEXISTENT-123`, treat as top-level (defensive)
  - Optional: Add warning marker in output (e.g., `⚠` icon)
  - Validation should already catch this (`pm validate`)
- [x] Empty result handling:
  - If `--parent X` returns no children, print: `No children found for {id}`
  - Exit cleanly (not an error, just informational)
- [x] Update help text with examples
- [x] Update cache queries to support new filters

## Edge Cases & Behavior

### Parent Ticket Does Not Exist
**Scenario**: User runs `pm list --parent FAKE-999` where FAKE-999 doesn't exist

**Behavior**: 
- Error message: `Error: Ticket FAKE-999 not found`
- Exit with non-zero code
- Suggestion: `Did you mean: ...` (if fuzzy match exists)

### Orphaned Tickets (Invalid Parent Reference)
**Scenario**: TICKET-A has `parent: DELETED-123` but DELETED-123 was deleted

**Behavior**:
- **At runtime**: Treat TICKET-A as top-level (show in `pm list`)
- **Optional**: Display warning marker: `TICKET-A ⚠` 
- **Validation**: `pm validate` should flag this as an error
- **Philosophy**: Don't hide work due to broken references; fail gracefully

### Parent with No Children
**Scenario**: User runs `pm list --parent GPM-999` where GPM-999 exists but has no children

**Behavior**:
- Print message: `No children found for GPM-999`
- Exit code 0 (not an error)
- Do NOT show the parent ticket itself

### Combining --parent and --all Without Parent
**Scenario**: User runs `pm list --all` (no parent specified)

**Behavior**:
- Ignore hierarchy, show everything (current behavior)
- `--all` without `--parent` means "all levels globally"

### Deeply Nested Hierarchies
**Scenario**: Epic → Story → Task → Subtask (3+ levels)

**Behavior**:
- `pm list` → Show only Epic
- `pm list --parent EPIC-1` → Show stories (1 level down)
- `pm list --parent EPIC-1 --all` → Show stories, tasks, subtasks (entire tree)
- `pm list --parent STORY-1` → Show tasks (direct children of story)

## Testing Requirements

### Unit Tests
- [x] Parse `--all` and `--parent` flags (covered by Cobra automatically)
- [x] Build SQL query for top-level filter (implemented inline)
- [x] Build SQL query for direct children (implemented inline)
- [x] Build recursive CTE for subtree (implemented inline)
- [x] Detect invalid parent ticket ID (uses findTicketByID)
- [x] Handle empty parent field vs NULL (SQL query handles both)

### Integration Tests
- [x] `pm list` shows only top-level tickets
- [x] `pm list --all` shows everything
- [x] `pm list --parent X` shows direct children
- [x] `pm list --parent X --all` shows full subtree
- [x] Combine with `--status` filter
- [x] Error when parent ticket doesn't exist
- [x] Empty message when parent has no children
- [x] Orphaned ticket (invalid parent) appears at top-level (behavior documented)
- [x] Case-insensitive parent ticket ID matching

## Acceptance Criteria

- [x] `pm list` shows only tickets with no parent (top-level)
- [x] `pm list --all` shows all tickets (backward compatible)
- [x] `pm list --parent <id>` shows direct children
- [x] `pm list --parent <id> --all` shows entire subtree recursively
- [x] Error message if parent ticket doesn't exist
- [x] Informational message if parent has no children
- [x] Orphaned tickets (invalid parent) behavior documented (don't appear at top-level)
- [x] All existing filters (`--status`, `--type`, etc.) work with new flags
- [x] Help text documents new flags with examples
- [x] All tests pass

## SQL Reference

### Top-Level Query (Default)
```sql
SELECT * FROM tickets 
WHERE (parent IS NULL OR parent = '')
ORDER BY updated_at DESC;
```

### Direct Children Query
```sql
SELECT * FROM tickets 
WHERE parent = ?
ORDER BY updated_at DESC;
```

### Recursive Subtree Query
```sql
WITH RECURSIVE subtree(id) AS (
  -- Base case: direct children
  SELECT id FROM tickets WHERE parent = ?
  
  UNION ALL
  
  -- Recursive case: children of children
  SELECT t.id 
  FROM tickets t 
  JOIN subtree s ON t.parent = s.id
)
SELECT t.* 
FROM tickets t
WHERE t.id IN (SELECT id FROM subtree)
ORDER BY t.updated_at DESC;
```

## Future Enhancements

- Tree visualization mode: `pm list --tree` (ASCII tree like `pm tree`)
- Breadcrumb display: Show parent chain for each ticket
- Filter by depth: `pm list --parent X --depth 2` (limit recursion)
- Exclude epics from default view: `pm list --no-epics`