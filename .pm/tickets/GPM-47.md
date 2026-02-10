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

### Database Schema Enhancement

**New Migration Required**: `000004_add_relationships_table`

Add a relationships table to cache for efficient reverse dependency lookups:
```sql
CREATE TABLE relationships (
  from_ticket TEXT NOT NULL,
  to_ticket TEXT NOT NULL,
  relationship_type TEXT NOT NULL,  -- 'depends-on', 'blocks', 'related'
  PRIMARY KEY (from_ticket, to_ticket, relationship_type)
);

CREATE INDEX idx_from ON relationships(from_ticket);
CREATE INDEX idx_to ON relationships(to_ticket);
CREATE INDEX idx_type ON relationships(relationship_type);
```

**Rationale**: Enables efficient "what depends on me" queries without scanning all tickets.

### Cache Sync Enhancement

Update `internal/cache/sync.go` to populate relationships table:
1. Clear relationships table before sync
2. For each ticket, read DependsOn/Blocks/Related arrays
3. Insert rows: `(ticket_id, dependency_id, 'depends-on')` for each dependency

### Query Logic

**For global view (no ID):**
```sql
SELECT DISTINCT t.id, t.title, t.status,
  GROUP_CONCAT(r.to_ticket || ':' || dep.title || ':' || dep.status) as blockers
FROM tickets t
JOIN relationships r ON r.from_ticket = t.id AND r.relationship_type = 'depends-on'
JOIN tickets dep ON dep.id = r.to_ticket
GROUP BY t.id
HAVING COUNT(CASE WHEN dep.status NOT IN (completed_states) THEN 1 END) > 0
ORDER BY t.updated_at DESC
```

**For specific ticket view (with ID):**
```sql
-- What this ticket depends on
SELECT to_ticket, title, status FROM relationships r
JOIN tickets t ON t.id = r.to_ticket
WHERE from_ticket = ? AND relationship_type = 'depends-on'

-- What depends on this ticket (reverse lookup)
SELECT from_ticket, title, status FROM relationships r
JOIN tickets t ON t.id = r.from_ticket
WHERE to_ticket = ? AND relationship_type = 'depends-on'
```

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

### Database Schema
- [ ] Create migration `000004_add_relationships_table.up.sql`
- [ ] Create migration `000004_add_relationships_table.down.sql`
- [ ] Verify migration auto-embeds in `internal/migrations/embed.go`

### Cache Sync
- [ ] Update `internal/cache/sync.go` to populate relationships table
- [ ] Extract relationships from ticket DependsOn/Blocks/Related arrays
- [ ] Clear and rebuild relationships table on each sync
- [ ] Test sync with existing tickets containing dependencies

### Command Implementation
- [ ] Create `cmd/pm/blocked.go` with cobra command structure
- [ ] Implement global view (no arguments):
  - [ ] Query tickets with unresolved dependencies using relationships table
  - [ ] Filter using workflow.IsCompleted() for state group checking
  - [ ] Format output with ticket ID, title, status, and blockers
  - [ ] Add summary statistics (X tickets blocked by Y dependencies)
- [ ] Implement specific ticket view (with ticket ID):
  - [ ] Query what this ticket depends on (forward lookup)
  - [ ] Query what depends on this ticket (reverse lookup via relationships table)
  - [ ] Display both directions with status indicators
  - [ ] Show summary line with counts
- [ ] Add color coding:
  - [ ] Green ✓ for completed dependencies
  - [ ] Red ✗ for unresolved dependencies
  - [ ] Red "MISSING" for referenced tickets that don't exist
- [ ] Add shell completion for ticket IDs
- [ ] Handle all edge cases gracefully

### Integration
- [ ] Register blocked command in `cmd/pm/main.go`
- [ ] Ensure lazy migration runs on first use
- [ ] Test with real GPM tickets (GPM-5, GPM-10, etc.)

### Testing
- [ ] Unit tests in `cmd/pm/blocked_test.go`
- [ ] Integration test in `integration_blocked_test.go`:
  - [ ] Create tickets with dependency chains
  - [ ] Verify global view shows only unresolved blocks
  - [ ] Verify specific view shows both directions
  - [ ] Test with completed dependencies (should be marked resolved)
  - [ ] Test with missing ticket references
  - [ ] Test with no blocked tickets (empty result)

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

- [ ] Migration 000004 creates relationships table with proper indexes
- [ ] Cache sync populates relationships table from ticket arrays
- [ ] `pm blocked` (no args) lists all tickets with unresolved dependencies
- [ ] `pm blocked <id>` shows what blocks the ticket and what it blocks
- [ ] Resolved dependencies are clearly indicated with green ✓
- [ ] Unresolved dependencies show with red ✗
- [ ] Missing ticket references show as "MISSING" in red
- [ ] Output is readable and well-formatted with proper truncation
- [ ] Color coding helps distinguish resolved/unresolved
- [ ] Shell completion works for ticket IDs
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Relationships table enables efficient reverse lookups

## Dependencies

- Requires GPM-45 (`pm link`) to be implemented first (to create dependency relationships) ✅ COMPLETED
- Uses state_groups from workflow.yaml (already implemented) ✅ AVAILABLE
- Uses cache database for efficient queries ✅ AVAILABLE
- Requires new migration (000004) for relationships table