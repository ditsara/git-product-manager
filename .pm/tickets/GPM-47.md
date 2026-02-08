---
id: GPM-47
title: "Implement pm blocked for dependency tracking"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 3  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: GPM-2  # Parent epic or story
depends_on: [GPM-45]  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: [relationships, cli, visualization]  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-08T14:43:40Z"
updated_at: "2026-02-08T14:43:40Z"
---

# Description

[Claude Sonnet 4.5]

Implement `pm blocked` command to show dependency and blocking relationships for tickets, helping identify bottlenecks in the workflow.

## Command Specifications

### `pm blocked [<id>]`

**Without ticket ID (global view):**
```bash
pm blocked
```
- Lists all tickets that have unresolved dependencies
- Shows what each ticket is blocked by
- Use case: "What work is currently blocked?"

**Example output:**
```
Blocked Tickets:

GPM-5: Implement Bad YAML validation
  Blocked by: GPM-10 (Database migrations)
  Status: backlog

GPM-9: Auto-recovery on database errors
  Blocked by: GPM-10 (Database migrations)
  Status: backlog

GPM-11: Implement pm repair command
  Blocked by: GPM-10 (Database migrations)
  Status: backlog

3 tickets blocked by 1 dependency
```

**With ticket ID (specific ticket view):**
```bash
pm blocked GPM-10
```
- Shows what blocks this ticket (its dependencies)
- Shows what this ticket blocks (tickets that depend on it)
- Use case: "What's the impact of completing/not completing this ticket?"

**Example output:**
```
GPM-10: Database migrations with golang-migrate

This ticket depends on:
  (none)

This ticket blocks:
  • GPM-5: Implement Bad YAML validation
  • GPM-9: Auto-recovery on database errors
  • GPM-11: Implement pm repair command
  • GPM-17: Cache metadata table
  • GPM-46: Cache sync symmetry validation

Status: done
Blocking 5 tickets (all unresolved)
```

## Implementation Approach

### Query Logic

**For global view (no ID):**
1. Read all tickets from cache
2. Filter tickets with non-empty `depends_on` arrays
3. For each dependency, check if the blocking ticket is in a "completed" state
4. Show only tickets with at least one unresolved dependency

**For specific ticket view (with ID):**
1. Read the specified ticket
2. Show its `depends_on` array (what blocks it)
3. Query all tickets to find which have this ticket in their `depends_on` array (what it blocks)
4. Display both directions with status information

### Using State Groups

Leverage the state_groups feature from workflow.yaml:
```go
func isResolved(ticketID string, workflow config.Workflow) bool {
    ticket := loadTicket(ticketID)
    return workflow.IsCompleted(ticket.Status)
}
```

A dependency is "unresolved" if the blocking ticket is NOT in a completed state.

## Implementation Steps

- [ ] Create `cmd/pm/blocked.go`
- [ ] Implement global view (no arguments):
  - [ ] Query all tickets with dependencies
  - [ ] Filter to show only unresolved blocks
  - [ ] Format output with ticket ID, title, and blockers
- [ ] Implement specific ticket view (with ticket ID):
  - [ ] Show what this ticket depends on
  - [ ] Query reverse: what depends on this ticket
  - [ ] Display summary statistics
- [ ] Add color coding:
  - [ ] Red for unresolved dependencies
  - [ ] Green for completed dependencies
- [ ] Add shell completion for ticket IDs
- [ ] Handle edge cases gracefully

## Examples

```bash
# See all blocked work
pm blocked

# Check impact of a specific ticket
pm blocked GPM-44
GPM-44: Reliability & Data Integrity (epic)

This ticket depends on:
  • GPM-10: Database migrations (done) ✓

This ticket blocks:
  (none - this is an epic, children are independent)

Status: backlog

# Check a completed ticket
pm blocked GPM-10
GPM-10: Database migrations

This ticket depends on:
  (none)

This ticket blocks:
  • GPM-5: Implement Bad YAML validation (backlog)
  • GPM-9: Auto-recovery on database errors (backlog)
  • GPM-11: Implement pm repair command (backlog)

Status: done ✓
Blocking 3 tickets
```

## Edge Cases

- **Ticket doesn't exist**: Error with helpful message
- **Ticket has no dependencies**: Show "This ticket depends on: (none)"
- **Nothing depends on this ticket**: Show "This ticket blocks: (none)"
- **Circular dependencies**: Detect and highlight (should already be prevented by validation)
- **Referenced ticket doesn't exist**: Show "MISSING" indicator
- **Empty database**: "No blocked tickets found"

## Display Formatting

- **Use colors** (if terminal supports):
  - Unresolved dependencies: Red ✗
  - Resolved dependencies: Green ✓
- **Indentation**: Use bullet points for clarity
- **Summary line**: Count of blocked tickets and blocking tickets
- **Truncate long titles**: Use consistent truncation (40 chars)

## Testing

- [ ] Unit test: Parse depends_on arrays
- [ ] Unit test: Determine if dependency is resolved (using state groups)
- [ ] Integration test: Global view shows only tickets with unresolved dependencies
- [ ] Integration test: Specific view shows both directions
- [ ] Integration test: Completed dependencies marked as resolved
- [ ] Test with no blocked tickets
- [ ] Test with missing dependency references

## Acceptance Criteria

- [ ] `pm blocked` (no args) lists all tickets with unresolved dependencies
- [ ] `pm blocked <id>` shows what blocks the ticket and what it blocks
- [ ] Resolved dependencies are clearly indicated
- [ ] Output is readable and well-formatted
- [ ] Color coding helps distinguish resolved/unresolved
- [ ] Shell completion works for ticket IDs
- [ ] All tests pass

## Dependencies

- Requires GPM-45 (`pm link`) to be implemented first (to create dependency relationships)
- Uses state_groups from workflow.yaml (already implemented)
- Uses cache database for efficient queries