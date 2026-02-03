---
id: GPM-22
title: "Implement pm assign shorthand command"
type: story
status: backlog
priority: low
points: 1

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-1"
depends_on: []
blocks: []
related: ["GPM-7"]  # GPM-7 implements pm edit --field

labels: [cli, ux, stage-2]
assignee: ""
created_at: "2026-02-03T14:22:54Z"
updated_at: "2026-02-03T14:22:54Z"
---

# Description

[Sonnet 4.5]

Implement `pm assign` as a convenience command for updating the assignee field without using the full `pm edit --field` syntax.

## User Story

As a team member, I want to quickly assign tickets to myself or colleagues without opening an editor or using verbose flags, so that I can manage task ownership efficiently.

## Solution Approach

Create a thin wrapper around the existing field update mechanism:

```bash
pm assign <ticket-id> <username>
```

This is equivalent to:
```bash
pm edit <ticket-id> --field assignee=<username>
```

## Implementation Steps

- [ ] Create `cmd/pm/assign.go` with cobra command
- [ ] Accept two arguments: ticket ID and username
- [ ] Validate ticket exists (use `findTicketByID()`)
- [ ] Validate username is non-empty
- [ ] Call existing assignee field update logic (reuse from `pm edit`)
- [ ] Auto-commit with message: `chore(pm): Assign {id} to {username}`
- [ ] Update `updated_at` timestamp
- [ ] Add command to root command registration

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

### Integration Tests
- [ ] Assign ticket, verify assignee field updated
- [ ] Verify `updated_at` timestamp changes
- [ ] Verify git commit created with correct message
- [ ] Case-insensitive ticket ID matching

## Acceptance Criteria

- [ ] `pm assign <id> <user>` updates assignee field
- [ ] Updated timestamp is automatically set
- [ ] Git commit created with descriptive message
- [ ] Error message for invalid ticket ID
- [ ] Error message for missing username
- [ ] Works with case-insensitive ticket IDs (e.g., `gpm-123`)
- [ ] All tests pass

## Edge Cases

- **Ticket already assigned**: Overwrite with new assignee (no warning)
- **Empty username**: Error "Username cannot be empty"
- **Assigning to same person**: No-op, but still update timestamp and commit
- **Special characters in username**: Accept as-is (validation is user's responsibility)
- **Clearing assignment**: Use `pm edit --field assignee=""` (not supported in assign command)

## Implementation Notes

This command is syntactic sugar for better UX. The actual field update logic should be shared with `pm edit --field` to avoid duplication. Consider extracting field update logic into `internal/ticket/fields.go` if not already done.