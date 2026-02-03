---
id: GPM-20
title: "Enhance pm show to display comments"
type: story
status: backlog
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
updated_at: "2026-02-03T14:22:54Z"
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

- [ ] Update `cmd/pm/show.go` to check for comment directory
- [ ] Read all `.md` files from `.pm/tickets/{id}/` directory
- [ ] Parse each comment file (YAML front-matter + markdown body)
- [ ] Sort comments by timestamp (chronological order)
- [ ] Format and display comments after ticket content
- [ ] Add separator line between ticket and comments
- [ ] Implement `--no-comments` flag
- [ ] Handle case where comment directory doesn't exist (no error)
- [ ] Handle case where comment directory is empty (show "No comments")

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
- [ ] Comment file parsing (YAML + markdown)
- [ ] Comment sorting (by timestamp)
- [ ] Comment formatting for display
- [ ] Handling of missing comment directory
- [ ] Handling of empty comment directory

### Integration Tests
- [ ] Show ticket with no comments (no error, no comments section)
- [ ] Show ticket with single comment
- [ ] Show ticket with multiple comments (verify chronological order)
- [ ] Show ticket with `--no-comments` flag (comments hidden)
- [ ] Show ticket with comments from multiple authors

## Acceptance Criteria

- [ ] `pm show <id>` displays all comments below ticket content
- [ ] Comments sorted chronologically (oldest first)
- [ ] Each comment shows author, timestamp, and text
- [ ] `--no-comments` flag suppresses comment display
- [ ] No error when ticket has no comments
- [ ] Clear separation between ticket content and comments
- [ ] Timestamps displayed in human-readable format (UTC)
- [ ] Multi-line comments displayed correctly
- [ ] All tests pass

## Edge Cases

- **No comment directory**: Silently skip comments section
- **Empty comment directory**: Show "No comments"
- **Malformed comment file**: Log warning, skip file, continue showing other comments
- **Very long comments**: Display fully (no truncation) - user can scroll terminal
- **Comments with special characters**: Markdown should render correctly