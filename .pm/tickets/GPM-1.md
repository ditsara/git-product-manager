---
id: GPM-1
title: "Stage 2: Collaboration and History"
type: epic
status: done
priority: high
points: 0
parent: ""
depends_on: []
blocks: ["GPM-19", "GPM-20", "GPM-21", "GPM-22", "GPM-23"]
related: []
labels: [collaboration, git-history, epic]
assignee: ""
created_at: 2026-02-02T03:54:36Z
updated_at: 2026-02-08T08:08:00Z
---

# Description

[Sonnet 4.5]

This epic tracks the implementation of Stage 2 features for team collaboration and auditing, centered around the conflict-free comment system and git history analysis.

## Child Stories

This epic has been broken down into the following user stories:

- **[GPM-19](GPM-19.md)**: Implement `pm comment` command (comment system)
- **[GPM-20](GPM-20.md)**: Enhance `pm show` to display comments
- **[GPM-21](GPM-21.md)**: Implement `pm history` for state change auditing
- **[GPM-22](GPM-22.md)**: Implement `pm assign` shorthand command
- **[GPM-23](GPM-23.md)**: Implement `pm edit-comment` to modify existing comments

## Objectives

- Enable multiple team members to comment on tickets without merge conflicts
- Provide audit trail of ticket state changes through git history
- Add convenient commands for ticket assignment

## Key Features

### 1. Comment System (`pm comment`)
Implement the full comment system with separate files to prevent merge conflicts:
- **Separate file per comment**: Store comments in `.pm/tickets/{TICKET-ID}/` directory
- **Filename format**: `{ISO8601-timestamp}-{author}.md`
- **Interactive mode**: Open editor for comment composition (respects `$VISUAL`, `$EDITOR`, etc.)
- **Direct mode**: `-m` flag for one-line comments without editor
- **Author override**: `--author` flag to specify commenter
- **YAML front-matter**: Include `author` and `timestamp` metadata (migrated from single timestamp field to created_at/updated_at for amendment support)
- **Comment amendment**: `pm comment --amend` to edit existing comments
- **Cache integration**: Update SQLite cache with new comments

### 2. Enhanced Ticket Display (`pm show`)
Update to integrate comment display:
- **Show all comments**: Read from `.pm/tickets/{id}/` directory
- **Chronological order**: Sort by timestamp
- **Format**: `@{author} ({timestamp}): {comment_text}`
- **Optional suppression**: `--no-comments` flag to hide comments

### 3. State Change History (`pm history`)
Implement auditing by parsing git history:
- **Extract state changes**: Parse git log for the ticket file
- **Show transitions**: Display `old_state → new_state` with author and date
- **Context**: Include commit messages for each change
- **Format**: `YYYY-MM-DD HH:MM  author  old_state → new_state`

### 4. Assignment Shorthand (`pm assign`)
Add convenient assignee command:
- **Syntax**: `pm assign <id> <user>`
- **Implementation**: Shorthand for `pm edit <id> --field assignee=<user>`

## Database Changes

### New Migration (000003)
- **Add `comments` table**:
  ```sql
  CREATE TABLE comments (
    ticket_id TEXT NOT NULL,
    author TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    filepath TEXT NOT NULL,
    PRIMARY KEY (ticket_id, timestamp, author)
  );
  CREATE INDEX idx_ticket_comments ON comments(ticket_id, timestamp);
  ```

### Cache Logic Updates
- **Index new comments**: Populate comments table on creation
- **Sync comments**: Include in cache sync logic
- **Query by ticket**: Support filtering comments by ticket ID

## Testing Requirements

### Unit Tests
- [x] Comment file parsing (YAML front-matter + markdown body)
- [x] Comment filename generation (ISO8601 format, filesystem-safe)
- [x] Git history parsing for state extraction
- [x] Comment directory creation and organization

### Integration Tests
- [x] Create comment via interactive mode (simulated editor)
- [x] Create comment via `-m` flag
- [x] View ticket with comments via `pm show`
- [x] Verify comments sorted chronologically
- [x] Test `--no-comments` flag
- [x] Verify `pm history` extracts state changes from git log
- [x] Test simultaneous comments from multiple "users" (no conflicts)
- [x] Test `pm comment --amend` to edit existing comments

## Acceptance Criteria

- [x] Users can add comments without opening the full ticket file
- [x] Multiple team members can comment simultaneously without merge conflicts
- [x] All comments are preserved in git history
- [x] State changes are auditable through `pm history`
- [x] Comment cache is automatically synced
- [x] All tests pass (unit + integration)
- [x] Comments can be amended with created_at/updated_at tracking
- [x] Assignee field can be updated via `pm assign` shorthand

## Implementation Checklist

- [x] **`pm comment`**: Implement full comment system (separate files, editor logic, `-m` flag, amendment)
- [x] **`pm show`**: Update to integrate comment display, including `--no-comments` flag
- [x] **`pm history`**: Implement state change auditing by parsing git history
- [x] **`pm assign`**: Add assignee shorthand command
- [x] **Database**: Add migration for `comments` table
- [x] **Database**: Update cache logic to index new comments
- [x] **Tests**: Unit tests for comment file parsing
- [x] **Tests**: Integration tests for creating, viewing, and listing comments
- [x] **Tests**: Tests for `pm history` command against sample git history
- [x] **Tests**: Tests for `pm assign` command with idempotency
- [x] **Tests**: Tests for `pm comment --amend` with created_at/updated_at tracking
