---
id: GPM-20
title: "Enhance pm show to display comments"
type: story
status: done
priority: high
points: 3

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-1"
depends_on: [GPM-19]
blocks: []
related: []

labels: [collaboration, stage-2]
assignee: ""
created_at: "2026-02-03T14:22:54Z"
updated_at: "2026-02-04T05:20:00Z"
---

# Description

[Sonnet 4.5]

Enhance the `pm show` command to display comments associated with a ticket, reading them from the comment directory and presenting them in chronological order.

## User Story

As a team member, I want to see all comments on a ticket when viewing it, so that I can understand the full discussion and context around the ticket.

## Solution Approach

When displaying a ticket with `pm show <id>`, automatically read and display all comments from `.pm/tickets/{id}/` directory:

1. **After showing ticket content**, add a "Comments" section
2. **Read all comment files** from the comment directory
3. **Sort chronologically** by timestamp (ascending - oldest first)
4. **Format for display**:
   ```
   Comments (3):
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   @alice (2026-02-03 14:30:00 UTC)
   This looks good! Just need to add tests.
   
   @bob (2026-02-03 15:00:00 UTC)
   Added tests in commit abc123.
   
   @alice (2026-02-03 15:15:00 UTC)
   LGTM! Merging now.
   ```

## Optional Suppression

Add `--no-comments` flag to hide comments when desired:
```bash
pm show GPM-123 --no-comments
```

## Implementation Steps

- [x] Update `cmd/pm/show.go` to check for comment directory
- [x] Read all `.md` files from `.pm/tickets/{id}/` directory
- [x] Parse each comment file (YAML front-matter + markdown body)
- [x] Sort comments by timestamp (chronological order)
- [x] Format and display comments after ticket content
- [x] Add separator line between ticket and comments
- [x] Implement `--no-comments` flag
- [x] Handle case where comment directory doesn't exist (no error)
- [x] Handle case where comment directory is empty (no section shown)

## Display Format

```
---
id: GPM-123
title: "Example ticket"
...
---

# Description

Ticket content here...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Comments (2):

@alice (2026-02-03 14:30:00 UTC)
First comment text here.
Can be multiple lines.

@bob (2026-02-03 15:00:00 UTC)
Second comment text here.
```

## Testing Requirements

### Unit Tests
- [x] Comment file parsing (YAML + markdown)
- [x] Comment sorting (by timestamp)
- [x] Comment formatting for display
- [x] Handling of missing comment directory
- [x] Handling of empty comment directory

### Integration Tests
- [x] Show ticket with no comments (no error, no comments section)
- [x] Show ticket with single comment
- [x] Show ticket with multiple comments (verify chronological order)
- [x] Show ticket with `--no-comments` flag (comments hidden)
- [x] Show ticket with comments from multiple authors

## Acceptance Criteria

- [x] `pm show <id>` displays all comments below ticket content
- [x] Comments sorted chronologically (oldest first)
- [x] Each comment shows author, timestamp, and text
- [x] `--no-comments` flag suppresses comment display
- [x] No error when ticket has no comments
- [x] Clear separation between ticket content and comments
- [x] Timestamps displayed in human-readable format (UTC)
- [x] Multi-line comments displayed correctly
- [x] All tests pass

## Edge Cases

- **No comment directory**: Silently skip comments section ✓
- **Empty comment directory**: Silently skip comments section ✓
- **Malformed comment file**: Handled gracefully ✓
- **Very long comments**: Display fully (no truncation) ✓
