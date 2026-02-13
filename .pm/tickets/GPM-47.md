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

**Scope**: The relationships table stores `depends-on` and `blocks` relationships only. The `related` field remains unidirectional (stored in source ticket only, not indexed). The `parent` field is NOT stored in relationships table as it's already indexed in the tickets table.

### Cache Sync Enhancement

Update `internal/cache/sync.go` to populate relationships table:
1. Clear relationships table before sync (at start of SyncCache transaction)
2. For each ticket, read DependsOn and Blocks arrays from parsed YAML
3. Insert rows for bidirectional relationships:
   - For each ID in `depends_on`: Insert `(ticket.ID, dependency_id, 'depends-on')`
   - For each ID in `blocks`: Insert `(ticket.ID, blocked_id, 'blocks')`
4. Note: Use relationship type `'depends-on'` (with hyphen) to match GPM-45 link command format

### Query Logic

**For global view (no ID):**

First, get completed states from workflow config in Go:
```go
workflow, _ := config.LoadWorkflow(workflowPath)
completedStates := workflow.GetCompletedStates() // e.g., ["done", "canceled"]
```

Then build SQL with dynamic placeholders for completed states:
```sql
-- Example with completed states: "done", "canceled"
SELECT DISTINCT t.id, t.title, t.status,
  GROUP_CONCAT(r.to_ticket || ':' || dep.title || ':' || dep.status) as blockers
FROM tickets t
JOIN relationships r ON r.from_ticket = t.id AND r.relationship_type = 'depends-on'
JOIN tickets dep ON dep.id = r.to_ticket
GROUP BY t.id
HAVING COUNT(CASE WHEN dep.status NOT IN ('done', 'canceled') THEN 1 END) > 0
ORDER BY t.updated_at DESC
```

**Implementation Note:** Build the `NOT IN` clause dynamically in Go based on `completedStates` array.

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
- [ ] Verify migration auto-embeds in `internal/migrations/embed.go` (using go:embed directive)
- [ ] Confirm relationship_type values use hyphenated format: 'depends-on', 'blocks' (matching pm link)

### Cache Sync
- [ ] Update `internal/cache/sync.go` to populate relationships table
- [ ] Add DELETE FROM relationships at start of transaction (alongside tickets/comments)
- [ ] Extract DependsOn and Blocks arrays from parsed ticket YAML
- [ ] Insert relationships using 'depends-on' and 'blocks' types (hyphenated format)
- [ ] Do NOT insert 'related' or 'parent' into relationships table (out of scope)
- [ ] Test sync with existing tickets containing dependencies (use GPM tickets as test data)

### Command Implementation
- [ ] Create `cmd/pm/blocked.go` with cobra command structure
- [ ] Implement global view (no arguments):
  - [ ] Load workflow config and get completed states using `workflow.GetCompletedStates()`
  - [ ] Build SQL query dynamically with completed states in NOT IN clause
  - [ ] Query tickets with unresolved dependencies using relationships table
  - [ ] Format output with ticket ID, title, status, and blockers
  - [ ] Add summary statistics (X tickets blocked by Y dependencies)
- [ ] Implement specific ticket view (with ticket ID):
  - [ ] Query what this ticket depends on (forward lookup from relationships table)
  - [ ] Query what depends on this ticket (reverse lookup: where to_ticket = ?)
  - [ ] Use workflow.IsCompleted() to mark resolved vs unresolved dependencies
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

---

## Dev Readiness Evaluation

**Evaluated by:** [Claude Sonnet 4.5]  
**Date:** 2026-02-13  
**Verdict:** ✅ **READY FOR IMPLEMENTATION** with minor clarifications below

### Summary

This ticket is well-specified and ready for implementation. All core requirements are clear, examples are comprehensive, and the technical approach is sound. The specification follows GPM best practices with detailed edge cases, acceptance criteria, and implementation steps.

### Strengths

1. **Clear Command Specifications**: Both command variants (with/without ticket ID) are precisely defined with expected output examples
2. **Complete Database Design**: Migration schema is fully specified with proper indexes
3. **Excellent Examples**: Multiple realistic examples using actual GPM tickets
4. **Edge Cases Well-Covered**: Missing tickets, empty results, circular dependencies all addressed
5. **Dependencies Verified**: GPM-45 is complete (status: done), state_groups infrastructure exists and is functional
6. **Technical Feasibility Confirmed**: 
   - Migration numbering is correct (next is 000004)
   - `workflow.IsCompleted()` method exists and works as specified
   - Cache sync pattern is established and understood
7. **Testability**: Acceptance criteria are specific and measurable

### Issues Found

#### Blocker Issues
**None** - Specification is implementable as-is.

#### Should-Fix Issues

**1. SQL Query Syntax Error**
- **Location:** Section "Query Logic" → "For global view"
- **Issue:** The SQL uses pseudo-code `completed_states` variable that doesn't exist in SQL
- **Current:**
  ```sql
  HAVING COUNT(CASE WHEN dep.status NOT IN (completed_states) THEN 1 END) > 0
  ```
- **Should be:**
  ```sql
  -- In Go code, get completed states from workflow config first:
  completedStates := workflow.GetCompletedStates() // Returns []string{"done", "canceled"}
  
  -- Then build SQL dynamically or use placeholders
  HAVING COUNT(CASE WHEN dep.status NOT IN ('done', 'canceled') THEN 1 END) > 0
  ```
- **Recommendation:** Add a note that the query needs to incorporate `workflow.GetCompletedStates()` results dynamically

**2. Relationship Type Naming Inconsistency**
- **Location:** Database schema and queries
- **Issue:** GPM-45 uses `depends-on` (with hyphen), but SQL examples use `depends_on` (underscore)
- **Impact:** Database will store either `depends-on` or `depends_on` - needs to be consistent
- **Recommendation:** Verify which format is actually stored by GPM-45 and use that consistently
- **Note:** Based on YAML field names (`depends_on`), likely stored as `depends-on` in relationships table

**3. Missing Detail: Relationship Table Population**
- **Location:** Section "Cache Sync Enhancement"
- **Issue:** Says "Update internal/cache/sync.go" but doesn't specify:
  - Should `related` field be indexed? (spec says 'depends-on', 'blocks', 'related')
  - What about `parent` field - should it go in relationships table?
- **Current behavior:** `pm list --parent GPM-2` already works, suggesting parent might not need relationships table
- **Recommendation:** Clarify if relationships table is ONLY for depends-on/blocks or also includes related/parent

#### Nice-to-Have Enhancements

**1. Color Indicator Symbols**
- The spec mentions "Red ✗" and "Green ✓" but doesn't specify terminal color codes
- **Suggestion:** Reference existing color usage in `pm list` or `pm show` for consistency

**2. Shell Completion Mention**
- Acceptance criteria mentions "Shell completion works for ticket IDs" but implementation steps don't detail this
- **Suggestion:** This likely just means using existing `completeTicketIDs` helper (same as other commands)

**3. Truncation Reference**
- Mentions "Use consistent truncation (40 chars)" but this is implemented inconsistently in the codebase
- **Suggestion:** If GPM has a standard truncate function now, reference it; otherwise this is fine as a guideline

### Recommendations for Implementation

1. **Before starting:** Verify the relationship_type values stored by GPM-45 (run `pm link` and check the database)
2. **SQL queries:** Use `workflow.GetCompletedStates()` to dynamically build the completed states list
3. **Relationship table scope:** Confirm whether to include `parent` and `related` or just `depends-on`/`blocks`
4. **Follow existing patterns:** The ticket correctly references `pm list` and `pm show` - follow their color/formatting conventions

### Implementation Checklist Status

All checklist items are actionable and clear:
- ✅ Database migration steps are specific
- ✅ Cache sync steps reference the right file
- ✅ Command implementation steps are detailed
- ✅ Test requirements are comprehensive

### Conclusion

This is one of the most thorough ticket specifications in the GPM project. The three "Should-Fix" issues are clarifications rather than blockers - an experienced developer could reasonably infer the correct behavior. However, making these explicit will prevent implementation drift.

**Recommended Action:** Update the SQL query example and clarify relationship table scope, then proceed with implementation.