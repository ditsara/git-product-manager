---
id: GPM-19
title: "Implement pm comment command for conflict-free commenting"
type: story
status: done
priority: high
points: 5

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-1"
depends_on: []
blocks: [GPM-20]
related: []

labels: [collaboration, stage-2]
assignee: ""
created_at: "2026-02-03T14:22:48Z"
updated_at: "2026-02-04T05:15:00Z"
---

# Description

[Sonnet 4.5]

Implement the `pm comment` command to allow team members to add comments to tickets without editing the main ticket file. This prevents merge conflicts when multiple people comment on the same ticket simultaneously.

## User Story

As a team member, I want to add comments to tickets without opening the full ticket file, so that I can provide feedback, ask questions, or share updates without risk of merge conflicts.

## Solution Approach

Store each comment as a separate file in `.pm/tickets/{TICKET-ID}/` directory:
- **Filename format**: `{ISO8601-timestamp}-{author}.md`
  - Timestamp uses hyphens instead of colons for filesystem compatibility
  - Example: `2026-02-03T14-30-00Z-alice.md`
  
- **File content**:
  ```markdown
  ---
  author: alice
  timestamp: 2026-02-03T14:30:00Z
  ---
  
  Comment text goes here...
  ```

## Command Modes

### Interactive Mode (Default)
Opens editor for comment composition:
- Uses standard fallback chain: `$VISUAL` → `$EDITOR` → `editor` → `nano` → `vi`
- Pre-populate with template showing ticket context (similar to `git commit`)
- Lines starting with `#` are ignored
- Abort if editor exits with empty content

### Direct Mode (`-m` flag)
Provide comment message directly without editor:
```bash
pm comment GPM-123 -m "Looks good to merge"
```

### Author Override
Default author from git config `user.name`, override with `--author`:
```bash
pm comment GPM-123 -m "Fixed" --author bob
```

## Implementation Steps

## Implementation Steps

- [x] Create `internal/ticket/comment.go` with comment file operations
- [x] Implement `parseCommentFile()` - read YAML + markdown
- [x] Implement `createCommentFile()` - generate filename, write file
- [x] Create `cmd/pm/comment.go` with cobra command
- [x] Implement interactive mode (editor selection and invocation)
- [x] Implement direct mode (`-m` flag)
- [x] Implement author detection (git config) and override (`--author`)
- [x] Update SQLite cache with new comment entry
- [x] Handle directory creation (`.pm/tickets/{id}/` may not exist)

## Database Changes

Add migration `000003_add_comments_table.up.sql`:
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

Update cache sync logic to index comments when scanning ticket directories.

## Testing Requirements

### Unit Tests
- [x] `parseCommentFile()` - Parse YAML front-matter + markdown
- [x] `createCommentFile()` - Generate correct filename format
- [x] Timestamp formatting (ISO8601 with hyphens)
- [x] Author detection from git config
- [x] Comment directory creation

### Integration Tests
- [x] Create comment via `-m` flag, verify file created
- [x] Create comment via `-m` flag, verify cache updated
- [x] Create comment with custom author, verify metadata
- [x] Create multiple comments on same ticket, verify no conflicts
- [x] Comment on non-existent ticket, verify error

## Acceptance Criteria

- [x] `pm comment GPM-1 -m "Test"` creates comment file with correct structure
- [x] Comment filename uses ISO8601 timestamp with hyphens (filesystem-safe)
- [x] YAML front-matter includes `author` and `timestamp`
- [x] Author defaults to git config `user.name`
- [x] `--author` flag overrides default author
- [x] Comment directory created automatically if it doesn't exist
- [x] Cache is updated with new comment entry
- [x] Multiple people can comment simultaneously without conflicts
- [x] All tests pass

## Edge Cases

- **Ticket doesn't exist**: Error message, exit gracefully ✓
- **No git user.name configured**: Prompt user to set or use `--author` ✓
- **Empty comment message**: Abort with message "Empty comment not saved" ✓
- **Filesystem permissions**: Handle directory creation failures gracefully ✓
- **Concurrent comments**: Same-second timestamps get unique filenames (append counter if needed) ✓
