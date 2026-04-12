---
assignee: ""
blocks: []
created_at: "2026-02-08T14:47:32Z"
depends_on: []
id: GPM-49
labels:
    - search
    - cache
    - cli
parent: GPM-2
points: 3
priority: medium
related: []
status: done
title: Implement pm search for full-text search
type: task
updated_at: "2026-04-12T10:25:17Z"
---


# Description

Implement `pm search` command for full-text search across ticket titles and body content using SQLite `LIKE` queries against the existing `tickets` table. No additional migration or virtual table required.

## Command Specification

### `pm search <query> [flags]`

**Search across**: Ticket ID, title, and markdown body content

**Default behaviour**: returns all matching tickets regardless of status.

**Flags** (mirroring `pm list`):
- `--status <value>` — filter by specific status
- `--active` — only todo and in-progress
- `--completed` — only done and canceled
- `--incomplete` — exclude done and canceled
- `--all` — explicit "no filter" (same as default, useful in scripts)

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

No new migration required. The `tickets` table already has `id`, `title`, and `body` columns. Search is implemented with parameterized `LIKE` queries with optional status `AND` clauses:

```sql
SELECT * FROM tickets
WHERE (id LIKE ? OR title LIKE ? OR body LIKE ?)
  [AND status IN (...)]
ORDER BY
  CASE WHEN id LIKE ? THEN 0
       WHEN title LIKE ? THEN 1
       ELSE 2
  END;
```

## Match Snippet

Extract and display a short excerpt around the first match in `body` using manual string slicing in Go (no SQLite `snippet()` needed).

## Implementation Steps

- [x] Create `cmd/pm/search.go`
- [x] Implement LIKE query with relevance ordering (ID > title > body)
- [x] Add match snippet extraction in Go
- [x] Add status filter flags (`--status`, `--active`, `--completed`, `--incomplete`, `--all`)
- [x] Write integration tests

## Acceptance Criteria

- [x] `pm search <query>` finds matching tickets across all statuses
- [x] Searches ticket ID, title, and body content
- [x] Results ordered by relevance (ID match first, then title, then body)
- [x] Match context snippet shown for body matches
- [x] Status filter flags narrow results correctly
- [x] All tests pass