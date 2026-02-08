---
assignee: ""
blocks: []
created_at: "2026-02-03T14:22:54Z"
depends_on: []
id: GPM-22
labels:
    - cli
    - ux
    - stage-2
parent: GPM-1
points: 1
priority: low
related:
    - GPM-7
status: todo
title: Implement pm assign shorthand command
type: story
updated_at: "2026-02-08T08:01:12Z"
---


# Description

[Claude Haiku 4.5]

Implement `pm assign` as a convenience command for updating the assignee field without using the full `pm edit --field` syntax.

## User Story

As a team member, I want to quickly assign tickets to myself or colleagues without opening an editor or using verbose flags, so that I can manage task ownership efficiently.

As a team manager or leader, I want to see ticket assignment in `pm history` so I can audit who was assigned to what and when.

## Solution Approach

Create a thin wrapper around the existing field update mechanism:

```bash
pm assign <ticket-id> <username>
```

This is equivalent to `pm edit <ticket-id> --field assignee=<username>`, but:
- Skips the editor entirely (direct field update)
- Does not commit (user is responsible for committing changes)
- Does not update timestamp if assignee already equals the requested user

**Note on pm history integration**: This requires extending `pm history` to show assignee changes, not just status changes (Option B). See GPM-28 for pm history enhancement.

## Implementation Steps

- [ ] Create `cmd/pm/assign.go` with cobra command
- [ ] Accept two arguments: ticket ID and username
- [ ] Validate ticket exists (use `findTicketByID()`)
- [ ] Validate username is non-empty
- [ ] Parse ticket file and check current assignee value
- [ ] If assignee already equals requested user: show message "Already assigned to {user}" and exit (no changes)
- [ ] If assignee differs: update the assignee field and `updated_at` timestamp
- [ ] Validate and write ticket file to disk
- [ ] Display confirmation message
- [ ] Add command to root command registration in main.go

## Command Signature

```bash
pm assign <id> <user>

Arguments:
  id         Ticket ID to assign (e.g., GPM-123)
  user       Username or email of assignee

Examples:
  pm assign GPM-123 alice
  pm assign GPM-456 bob@example.com
```

## Testing Requirements

### Unit Tests
- [ ] Test assignee field update
- [ ] Test ticket ID validation (invalid ID)
- [ ] Test empty username (error handling)
- [ ] Test idempotent assignment (same user, no changes)

### Integration Tests
- [ ] Assign ticket, verify assignee field updated
- [ ] Verify `updated_at` timestamp changes
- [ ] Verify confirmation message displayed
- [ ] Verify no changes when assigning to same user
- [ ] Verify message printed when assignment is already current
- [ ] Case-insensitive ticket ID matching

## Acceptance Criteria

- [ ] `pm assign <id> <user>` updates assignee field
- [ ] Updated timestamp is automatically set
- [ ] Confirmation message printed (e.g., "Assigned GPM-123 to alice")
- [ ] Error message for invalid ticket ID
- [ ] Error message for missing username
- [ ] Works with case-insensitive ticket IDs (e.g., `gpm-123`)
- [ ] No changes made if assignee is already that user
- [ ] All tests pass

## Edge Cases

- **Ticket already assigned**: Print "Already assigned to {user}" and exit (no changes)
- **Empty username**: Error "Username cannot be empty"
- **Special characters in username**: Accept as-is (validation is user's responsibility)
- **Clearing assignment**: Use `pm edit --field assignee=""` (not supported in assign command)

## Implementation Notes

**Function Reuse**: Call the field update logic from `cmd/pm/edit.go`. The key is to:
1. Parse YAML front-matter
2. Check if current assignee equals requested user (idempotent check)
3. Update the `assignee` field if different
4. Update the `updated_at` timestamp if changed
5. Validate the result
6. Write back to disk

Do **not** commit changes; the user is responsible for committing via `git add` and `git commit`.

**pm history Integration**: Once this is implemented, create GPM-28 to extend `pm history` to show assignee (and other field) changes via git diff parsing, not just status changes.