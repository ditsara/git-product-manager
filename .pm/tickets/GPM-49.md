---
id: GPM-49
title: "Implement pm search for full-text search"
type: task
status: backlog
priority: medium
points: 5

parent: GPM-2
depends_on: []
blocks: []
related: []

labels: [search, cache, cli]
assignee: ""
created_at: "2026-02-08T14:47:32Z"
updated_at: "2026-02-08T14:47:32Z"
---

# Description

[Claude Sonnet 4.5]

Implement `pm search` command for efficient full-text search across ticket titles and content using SQLite FTS5.

## Command Specification

### `pm search <query> [--limit N]`

**Search across**: Ticket ID, title, and markdown body content

**Example output**:
```
Search results for "authentication" (3 matches):

GPM-26: Add visual indicators for ticket hierarchy
  Type: story | Status: done | Points: 2
  Match: ...visual hierarchy indicators in ticket lists...

GPM-1: Stage 2: Collaboration & History
  Type: epic | Status: backlog
  Match: ...Complete the vision by adding powerful relationship...
```

## Database Schema

### New Migration (000004)

Add `tickets_fts` virtual table for full-text search:

```sql
CREATE VIRTUAL TABLE tickets_fts USING fts5(
  id,
  title,
  body,
  content='tickets',
  content_rowid='rowid'
);
```

### Cache Logic Updates

- **Update FTS during sync**: When cache syncs from filesystem, update FTS table
- **Index all changes**: Every ticket create/update rebuilds FTS entries
- **Query optimization**: Use FTS ranking to sort results by relevance

## Implementation Steps

- [ ] Add `tickets_fts` table in migration
- [ ] Create `cmd/pm/search.go`
- [ ] Implement FTS query execution
- [ ] Add result pagination with `--limit` flag
- [ ] Update cache sync to rebuild FTS on every update

## Acceptance Criteria

- [ ] `pm search <query>` finds matching tickets
- [ ] Searches ticket ID, title, and body content
- [ ] Results sorted by relevance
- [ ] Search performance acceptable (<100ms for 1000 tickets)
- [ ] All tests pass
