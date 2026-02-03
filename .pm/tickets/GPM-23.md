---
id: GPM-23
title: "Implement pm edit-comment to modify existing comments"
type: story
status: backlog
priority: medium
points: 2

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-1"
depends_on: ["GPM-19"]  # Need comment system first
blocks: []
related: ["GPM-20"]  # Related to comment display

labels: [comments, editing, stage-2]
assignee: ""
created_at: "2026-02-03T14:41:14Z"
updated_at: "2026-02-03T14:41:14Z"
---

# Description

[Sonnet 4.5]

Implement the `pm edit-comment` command to allow users to modify or delete existing comments after they've been created.

## User Story

As a team member, I want to edit or delete my comments after posting them, so that I can fix typos, update information, or remove outdated remarks.

A comment which has been edited will show the editor and time (from git history) when viewed with `pm show`.

## Solution Approach

Allow editing comment files directly:

1. **List comments** for a ticket to select which one to edit
2. **Open comment file** in `$EDITOR` for modification
3. **Update timestamp** in YAML front-matter to track edits
4. **Auto-commit** with descriptive message
5. **Support deletion** via `--delete` flag

## Command Signature

```bash
pm edit-comment <ticket-id> [--author AUTHOR] [--timestamp TIMESTAMP] [--delete]

Arguments:
  ticket-id      Ticket ID containing the comment

Options:
  --author       Filter by comment author (default: current user)
  --timestamp    Specific comment timestamp to edit (ISO8601 format)
  --delete       Delete the comment instead of editing
  --list         List all comments for selection (interactive mode)

Examples:
  pm edit-comment GPM-123                          # Interactive: list and select comment
  pm edit-comment GPM-123 --author alice           # Edit most recent comment by alice
  pm edit-comment GPM-123 --timestamp 2026-02-01T14-30-00Z-alice  # Edit specific comment
  pm edit-comment GPM-123 --delete --author bob    # Delete most recent comment by bob
```

## Implementation Steps

- [ ] Create `cmd/pm/edit_comment.go` with cobra command
- [ ] Implement comment file discovery in `.pm/tickets/{id}/` directory
- [ ] Interactive mode: list comments and prompt for selection
  - Display: `[1] @alice (2026-02-01 14:30): First line of comment...`
  - Read user input for selection (1, 2, 3, etc.)
- [ ] Direct mode: Find specific comment by author and/or timestamp
  - If only `--author` provided, select most recent comment by that author
  - If `--timestamp` provided, find exact match
- [ ] Edit mode: Open comment file in `$EDITOR`
  - Use standard fallback chain: `$VISUAL` → `$EDITOR` → `editor` → `nano` → `vi`
  - Wait for editor to close
  - Update YAML front-matter: add `edited_at` field with current timestamp
- [ ] Delete mode: Remove comment file with confirmation
  - Prompt: "Delete comment by {author} from {timestamp}? [y/N]"
  - Remove file from filesystem
  - Update SQLite cache (remove from comments table)
- [ ] Validation: Ensure comment file exists before attempting edit/delete
- [ ] Auto-commit after edit/delete with message:
  - Edit: `comment(pm): Edit comment on {ticket-id} by {author}`
  - Delete: `comment(pm): Delete comment on {ticket-id} by {author}`
- [ ] Update cache after modification

## Comment File Format After Edit

```markdown
---
author: alice
timestamp: 2026-02-01T14:30:00Z
edited_at: 2026-02-03T16:45:00Z  # Added on edit
---

Updated comment text here.
```

## Interactive Mode Flow

```bash
$ pm edit-comment GPM-123

Comments on GPM-123:
[1] @alice (2026-02-01 14:30): Should we use OAuth2 library...
[2] @bob   (2026-02-02 09:15): I agree with alice, let's...
[3] @alice (2026-02-03 11:00): Update: I found a good library...

Select comment to edit [1-3] (or 'q' to cancel): 3

[Opens editor with comment #3]
```

## Testing Requirements

### Unit Tests
- [ ] Parse comment filename format (extract author and timestamp)
- [ ] Find comments by author
- [ ] Find comments by timestamp
- [ ] Sort comments chronologically
- [ ] Handle edited_at field in YAML front-matter

### Integration Tests
- [ ] Create comment, then edit it
- [ ] Verify `edited_at` timestamp added to front-matter
- [ ] Edit comment interactively (simulate user selection)
- [ ] Edit comment with `--author` flag
- [ ] Edit comment with `--timestamp` flag
- [ ] Delete comment with confirmation
- [ ] Delete comment bypassing confirmation (testing only)
- [ ] Verify cache updated after edit/delete
- [ ] Error handling: ticket not found
- [ ] Error handling: no comments exist
- [ ] Error handling: comment file not found

## Acceptance Criteria

- [ ] `pm edit-comment <id>` opens interactive comment selection
- [ ] Selected comment opens in `$EDITOR`
- [ ] `edited_at` timestamp added to front-matter after edit
- [ ] `--delete` flag removes comment file
- [ ] Delete prompts for confirmation (y/N)
- [ ] Git commit created after edit/delete
- [ ] Cache synchronized after modification
- [ ] Works with case-insensitive ticket IDs
- [ ] Error message if ticket has no comments
- [ ] Error message if specified comment not found
- [ ] All tests pass

## Edge Cases

- **No comments exist**: Error "No comments found for {ticket-id}"
- **Multiple edits**: Update `edited_at` timestamp on each edit (don't track edit history)
- **Author doesn't match**: If `--author bob` but bob has no comments, error "No comments by bob"
- **Ambiguous timestamp**: If multiple comments at same second, use filename as tiebreaker
- **Comment file manually deleted**: Gracefully handle missing file ("Comment file not found")
- **Editor aborted**: If editor returns non-zero or file unchanged, cancel operation
- **Permission to edit others' comments**: Allow editing any comment (rely on git history for audit)

## Security Considerations

- **Audit trail**: Git history preserves original comment even after edits
- **No permission system**: Any user can edit any comment (trust-based, like git)
- **Edited marker**: `edited_at` field makes edits visible to readers
- **Deletion is permanent**: Deleted comments removed from filesystem but preserved in git history

## Future Enhancements

- `--no-confirm` flag to skip delete confirmation
- Display edit history from git log
- Inline edit mode: `pm edit-comment <id> -m "New text"`
- Restore deleted comments from git history