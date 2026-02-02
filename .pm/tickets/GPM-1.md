---
id: GPM-1
title: "Stage 2: Collaboration and History"
type: epic
status: backlog
priority: high
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: [collaboration, git-history]
assignee: ""
created_at: 2026-02-02T03:54:36Z
updated_at: 2026-02-02T03:54:36Z
---

# Description

This stage introduces features for team collaboration and auditing, centered around the conflict-free comment system and git history analysis.

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
- **YAML front-matter**: Include `author` and `timestamp` metadata
- **Auto-commit**: Commit each comment with descriptive message
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
- [ ] Comment file parsing (YAML front-matter + markdown body)
- [ ] Comment filename generation (ISO8601 format, filesystem-safe)
- [ ] Git history parsing for state extraction
- [ ] Comment directory creation and organization

### Integration Tests
- [ ] Create comment via interactive mode (simulated editor)
- [ ] Create comment via `-m` flag
- [ ] View ticket with comments via `pm show`
- [ ] Verify comments sorted chronologically
- [ ] Test `--no-comments` flag
- [ ] Verify `pm history` extracts state changes from git log
- [ ] Test simultaneous comments from multiple "users" (no conflicts)

## Acceptance Criteria

- [ ] Users can add comments without opening the full ticket file
- [ ] Multiple team members can comment simultaneously without merge conflicts
- [ ] All comments are preserved in git history
- [ ] State changes are auditable through `pm history`
- [ ] Comment cache is automatically synced
- [ ] All tests pass (unit + integration)
- [ ] Documentation updated with new commands

## Implementation Checklist

- [ ] **`pm comment`**: Implement full comment system (separate files, editor logic, `-m` flag)
- [ ] **`pm show`**: Update to integrate comment display, including `--no-comments` flag
- [ ] **`pm history`**: Implement state change auditing by parsing git history
- [ ] **`pm assign`**: Add assignee shorthand command
- [ ] **Database**: Add migration for `comments` table
- [ ] **Database**: Update cache logic to index new comments
- [ ] **Tests**: Unit tests for comment file parsing
- [ ] **Tests**: Integration tests for creating, viewing, and listing comments
- [ ] **Tests**: Tests for `pm history` command against sample git history
