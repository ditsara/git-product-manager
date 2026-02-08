---
id: GPM-50
title: "Enhance pm list with relationship-aware filtering"
type: task
status: backlog
priority: medium
points: 3

parent: GPM-2
depends_on: [GPM-45]
blocks: []
related: []

labels: [relationships, cli, filtering]
assignee: ""
created_at: "2026-02-08T14:55:20Z"
updated_at: "2026-02-08T14:55:20Z"
---

# Description

[Claude Sonnet 4.5]

Enhance `pm list` command with relationship-aware filtering and database support for efficient querying of related tickets.

## Database Schema

### New Migration (000004)

Add `relationships` table for efficient querying of ticket relationships:

```sql
CREATE TABLE relationships (
  from_ticket TEXT NOT NULL,
  to_ticket TEXT NOT NULL,
  relationship_type TEXT NOT NULL,
  PRIMARY KEY (from_ticket, to_ticket, relationship_type)
);

CREATE INDEX idx_from ON relationships(from_ticket);
CREATE INDEX idx_to ON relationships(to_ticket);
```

**Rows stored for each relationship**:
- `depends_on`: Stored in both directions (bidirectional)
- `blocks`: Stored in both directions (bidirectional)
- `related`: Stored only in forward direction (unidirectional)

### Cache Logic Updates

- **Index relationships on sync**: When syncing tickets, parse all relationship arrays and populate relationships table
- **Maintain symmetry**: Ensure depends_on/blocks pairs are stored bidirectionally
- **Update on link/unlink**: When user runs `pm link` or `pm unlink`, update both ticket files AND relationships table

## Enhanced Filtering Flags

### New Flags to Add

**Filter by blocking relationships:**
```bash
pm list --depends-on <id>
# Show all tickets that depend on <id>

pm list --blocks <id> 
# Show all tickets that <id> blocks

pm list --related <id>
# Show all tickets related to <id>
```

**Example usage:**
```bash
# What work depends on GPM-10 (Database migrations)?
pm list --depends-on GPM-10

# What is GPM-44 blocking?
pm list --blocks GPM-44

# What's related to GPM-5?
pm list --related GPM-5
```

**Combining filters:**
```bash
# Show active work that depends on GPM-10
pm list --depends-on GPM-10 --active

# Show child tasks by status
pm list --parent GPM-44 --status backlog
```

## Implementation Steps

- [ ] Create migration 000004 with relationships table
- [ ] Update cache sync logic to populate relationships table
- [ ] Update `pm link` command to also update relationships table
- [ ] Update `pm unlink` command to also update relationships table
- [ ] Add new filter flags to `pm list`:
  - [ ] `--depends-on` flag
  - [ ] `--blocks` flag
  - [ ] `--related` flag
- [ ] Add shell completion for relationship filters
- [ ] Test filter combination logic

## Acceptance Criteria

- [ ] Relationships table stores and queries efficiently
- [ ] `--depends-on` returns all tickets depending on target
- [ ] `--blocks` returns all tickets blocked by target
- [ ] `--related` returns related tickets bidirectionally
- [ ] Filters combine with AND logic
- [ ] Works in combination with existing filters
- [ ] All tests pass

## Dependencies

- Requires GPM-45 (`pm link`) implementation
- Requires GPM-46 (symmetry validation) for data integrity
