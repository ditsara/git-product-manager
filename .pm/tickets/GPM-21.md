---
id: GPM-21
title: "Implement pm history for state change auditing"
type: story
status: backlog
priority: medium
points: 3

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-1"
depends_on: []
blocks: []
related: []

labels: [git-history, auditing, stage-2]
assignee: ""
created_at: "2026-02-03T14:22:54Z"
updated_at: "2026-02-08T00:12:00Z"
---

# Description

[Sonnet 4.5]

Implement the `pm history` command to show an audit trail of state changes for a ticket by parsing the git history and extracting status field modifications.

## User Story

As a project manager or team member, I want to see the complete history of state changes for a ticket, so that I can understand when and by whom the ticket moved through different workflow stages.

## Solution Approach

Parse git history for the ticket file and extract state transitions:

1. **Run git log** for the specific ticket file: `git log --follow .pm/tickets/{id}.md`
2. **For each commit**: Parse the diff to extract `status` field changes
3. **Build timeline** of state transitions
4. **Display** with author, timestamp, and optional commit message

Created status log: parse the first commit that adds the file and read status from that file content (not a diff).

## Display Format

```bash
$ pm history GPM-123

State Change History for GPM-123:

2026-02-03 09:00  alice     Created (status: backlog)
2026-02-03 10:30  bob       backlog → todo
    commit: Prioritized for next sprint
2026-02-03 14:00  alice     todo → in-progress
    commit: Started implementation
2026-02-03 16:45  alice     in-progress → done
    commit: Completed feature and tests
```

## Implementation Steps

- [ ] Implement `getGitLog(ticketPath)` in `cmd/pm/history.go` (shell out to git)
- [ ] Implement `parseCommitDiff()` - extract status field from diff
- [ ] Create `cmd/pm/history.go` with cobra command
- [ ] Resolve ticket path with `findTicketByID()` (case-insensitive)
- [ ] Build timeline of state changes with author and timestamp
- [ ] Format and display history (chronological order, oldest first)
- [ ] Include commit messages for context
- [ ] Handle special case: ticket creation (initial status)
- [ ] Handle case: no git history (ticket not committed)
- [ ] Handle case: no state changes (status never modified)
- [ ] Handle errors: git not installed or not a git repo

## Git Operations

### Chosen Option: Shell Out (Simpler)

Start with shell out for simplicity, can refactor to go-git later.

```go
cmd := exec.Command("git", "log", "--follow", "--format=%H|%an|%at|%s", ticketPath)
output, _ := cmd.Output()
// Parse output and for each commit, get diff
```

### Alternative: go-git Library (Pure Go)

We did not choose to do this; just for reference.

```go
import "github.com/go-git/go-git/v5"
repo, _ := git.PlainOpen(".")
log, _ := repo.Log(&git.LogOptions{FileName: ticketPath})
// Iterate through commits
```

## Parsing Strategy

For each commit:
1. Get the commit diff: `git show {commit_hash} -- {ticketPath}`
2. Look for lines matching: `-status: <old_value>` and `+status: <new_value>` (trim whitespace and optional quotes)
3. Extract old and new status values
4. Build transition record

## Testing Requirements

### Unit Tests
- [ ] Parse git log output format
- [ ] Extract status field from diff
- [ ] Build state change timeline
- [ ] Handle creation (no previous status)
- [ ] Handle no state changes

### Integration Tests  
- [ ] Create ticket, commit, move through states, verify history
- [ ] Ticket with single state change
- [ ] Ticket with multiple state changes
- [ ] Ticket never committed (error handling)
- [ ] Ticket never moved (only shows creation)
- [ ] Error handling: git not installed / not a git repo

## Acceptance Criteria

- [ ] `pm history <id>` shows all state changes chronologically
- [ ] Each entry shows date, author, and state transition
- [ ] Commit messages included for context (commit subject line only)
- [ ] Creation event shown (initial status)
- [ ] Works with tickets that have never changed status
- [ ] Error message if ticket file not in git history
- [ ] Timestamps displayed in human-readable format
- [ ] All tests pass

## Edge Cases

- **Ticket file not committed**: Error message "Ticket not in git history"
- **No status changes**: Show only creation event
- **Ticket file moved/renamed**: Use `--follow` to track history
- **Non-standard diff format**: Gracefully skip unparseable commits
- **Very long history**: Handle later; for now user can pipe through `less`
- **Git not available / not a repo**: Print error message and exit