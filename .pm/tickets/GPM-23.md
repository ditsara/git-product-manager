---
assignee: ""
blocks: []
created_at: "2026-02-03T14:41:14Z"
depends_on:
    - GPM-19
id: GPM-23
labels:
    - comments
    - editing
    - stage-2
parent: GPM-1
points: 2
priority: medium
related:
    - GPM-20
status: done
title: Implement pm comment --amend to modify existing comments
type: story
updated_at: "2026-02-08T00:10:50Z"
---


# Description

Implement the `pm comment --amend` command to allow users to modify existing comments after they've been created.

## User Story

As a team member, I want to edit my comments after posting them, so that I can fix typos, update information, or remove outdated remarks.

A comment which has been edited will show the editor and time when viewed with `pm show`. For simplicity, for now we will use the `updated_at` field.

## Solution Approach

Allow editing comment files directly:

1. **List comments** for a ticket to select which one to edit
2. **Open comment file** in `$EDITOR` for modification
3. **Update `updated_at`** in YAML front-matter to track edits

## Command Signature

```bash
pm comment <ticket-id> --amend [--author AUTHOR] [--timestamp TIMESTAMP]

Arguments:
  ticket-id      Ticket ID containing the comment

Options:
  --amend        Edit an existing comment, rather than create a new one (default: most recent comment)
  --author       Filter by comment author (default: current user)
  --timestamp    Specific comment timestamp to edit (ISO8601 format)

Examples:
  pm comment GPM-123 --amend                      # Interactive: list and select comment
  pm comment GPM-123 --amend --author alice       # Edit most recent comment by alice
  pm comment GPM-123 --amend --timestamp 2026-02-01T14-30-00Z  # Edit specific comment
```

## Implementation Steps

- [x] Integrate `--amend` flag into existing `cmd/pm/comment.go` command
- [x] Implement comment file discovery in `.pm/tickets/{id}/` directory
- [x] Change comments YAML front matter from `timestamp` to `created_at` and `updated_at`
- [x] Direct mode: Find specific comment by author and/or timestamp
  - If only `--author` provided, select most recent comment by that author
  - If `--timestamp` provided, find exact match
- [x] Edit mode: Open comment file in `$EDITOR`
  - Use standard fallback chain: `$VISUAL` → `$EDITOR` → `editor` → `nano` → `vi`
  - Wait for editor to close
  - Update YAML front-matter: set `updated_at` to current timestamp
- [x] Validation: Ensure comment file exists before attempting edit
- [x] No auto-commit (consistent with existing `pm comment` behavior)
- [x] Update cache after modification

## Comment File Format After Edit

```markdown
---
author: alice
created_at: 2026-02-01T14:30:00Z
updated_at: 2026-02-03T16:45:00Z  # Updated on edit
---

Updated comment text here.
```

## Interactive Mode Flow

```bash
$ pm comment GPM-123 --amend

Comments on GPM-123:
[1] @alice (2026-02-01 14:30): Should we use OAuth2 library...
[2] @bob   (2026-02-02 09:15): I agree with alice, let's...
[3] @alice (2026-02-03 11:00): Update: I found a good library...

Select comment to edit [1-3] (or 'q' to cancel): 3

[Opens editor with comment #3]
```

## Testing Requirements

### Unit Tests
- [x] Parse comment filename format (extract author and timestamp)
- [x] Find comments by author (covered by integration tests)
- [x] Find comments by timestamp (covered by integration tests)
- [x] Sort comments chronologically
- [x] Handle created_at/updated_at fields in YAML front-matter

### Integration Tests
- [x] Create comment, then edit it with `--amend`
- [x] Verify `updated_at` timestamp updated in front-matter
- [x] Edit comment interactively (simulate user selection)
- [x] Edit comment with `--author` flag
- [x] Edit comment with `--timestamp` flag
- [x] Verify cache updated after edit
- [x] Error handling: ticket not found
- [x] Error handling: no comments exist
- [x] Error handling: comment file not found
- [x] Verify distinguishing between `pm comment` (new comment) vs `pm comment --amend` (edit comment)

## Acceptance Criteria

- [x] `pm comment <id> --amend` opens interactive comment selection
- [x] Selected comment opens in `$EDITOR`
- [x] `updated_at` timestamp updated in front-matter after edit
- [x] No auto-commit (consistent with existing `pm comment` behavior)
- [x] Cache synchronized after modification
- [x] Works with case-insensitive ticket IDs
- [x] Error message if ticket has no comments
- [x] Error message if specified comment not found
- [x] `--amend` flag distinguishes from creating new comments
- [x] All tests pass

## Edge Cases

- **No comments exist**: Error "No comments found for {ticket-id}"
- **Multiple edits**: Update `updated_at` timestamp on each edit (don't track full edit history)
- **Author doesn't match**: If `--author bob` but bob has no comments, error "No comments by bob"
- **Ambiguous timestamp**: If multiple comments at same second, use filename as tiebreaker
- **Comment file manually deleted**: Gracefully handle missing file ("Comment file not found")
- **Editor aborted**: If editor returns non-zero or file unchanged, cancel operation (preserve original)
- **Permission to edit others' comments**: Allow editing any comment (rely on git history for audit)
- **Content removal**: Users can remove comment content if needed (leaves `updated_at` timestamp showing it was modified)

## Security Considerations

- **Audit trail**: Git history preserves original comment even after edits
- **No permission system**: Any user can edit any comment (trust-based, like git)
- **Edited marker**: `updated_at` field makes edits visible to readers
- **Transparency**: Comment modifications tracked in git history

## Future Enhancements

- Display edit history from git log (show `updated_at` timestamps when viewing)
- Inline edit mode: `pm comment <id> --amend -m "New text"` (direct update without editor)
- View comment edit diffs: `pm show <id> --comment-diff` to show what changed