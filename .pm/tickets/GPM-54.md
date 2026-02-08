---
id: GPM-54
title: "Add milestone field to ticket schema"
type: task
status: backlog
priority: medium
points: 3

parent: GPM-14
depends_on: [GPM-52]
blocks: [GPM-55, GPM-56]
related: []

labels: [milestone, schema]
assignee: ""
created_at: "2026-02-08T15:04:51Z"
updated_at: "2026-02-08T15:04:51Z"
---

# Description

[Claude Haiku 4.5]

**Task:** Extend the ticket YAML schema to support milestone associations and enable milestone assignment via `pm edit`.

## Overview

This task integrates milestones into the existing ticket system by adding a `milestones` field to ticket metadata. A ticket can belong to multiple milestones, enabling flexible grouping across overlapping milestone scopes.

## Implementation Steps

- [ ] Add `milestones: []` field to ticket YAML schema (array of milestone IDs)
- [ ] Update ticket templates (story, task, bug, epic) to include milestones field
- [ ] Create migration 000005b to alter tickets table:
  ```sql
  ALTER TABLE tickets ADD COLUMN milestones TEXT;  -- JSON array as text
  ```
  - Note: SQLite doesn't support JSON natively in older versions; store as comma-separated IDs or JSON array string
  - Decision: Store as comma-separated IDs for simplicity ("v1-0-release,sprint-3")
- [ ] Implement milestone assignment in `pm edit`:
  - `pm edit GPM-1 --field milestones=v1-0-release` (single)
  - `pm edit GPM-1 --field milestones=v1-0-release,sprint-3` (multiple, comma-separated)
  - Parse comma-separated list, validate each milestone ID exists
  - Show error: "Milestone not found: invalid-id"
- [ ] Implement milestone removal:
  - `pm edit GPM-1 --field milestones=` (clears all milestones)
- [ ] Add validation logic:
  - When setting milestones on a ticket, validate all referenced milestone IDs exist
  - Load milestone IDs from filesystem scan of `.pm/milestones/`
  - Provide helpful error message if milestone doesn't exist
  - Validation runs before file write
- [ ] Update SQLite cache:
  - Add milestones column to tickets table
  - Index on milestones for fast filtering: `CREATE INDEX idx_ticket_milestones ON tickets(milestones);`
  - Sync ticket milestones data from filesystem
- [ ] Update validator (`pm validate`):
  - Check that all milestone IDs referenced in tickets exist in `.pm/milestones/`
  - List any orphaned or broken references
  - Suggest: "Run `pm edit {ticket-id} --field milestones=` to remove"

## Acceptance Criteria

- [ ] Ticket YAML includes `milestones: []` field
- [ ] `pm edit GPM-1 --field milestones=v1-0-release` adds milestone to ticket
- [ ] `pm edit GPM-1 --field milestones=` clears milestones
- [ ] Validation rejects non-existent milestone IDs with helpful error
- [ ] SQLite cache stores and indexes milestones for fast queries
- [ ] `pm validate` detects orphaned milestone references
- [ ] Integration test: create milestone → assign to ticket → verify in cache
- [ ] Backward compatibility: existing tickets without milestones field work fine

## Code Output

- Updated `internal/ticket/ticket.go`: Add milestones field to Ticket struct
- Updated `internal/ticket/validator.go`: Add milestone validation logic
- Updated `cmd/pm/edit.go`: Handle --field milestones assignment
- Updated `.pm/config/templates/*.md`: Add milestones field
- Updated `internal/cache/sync.go`: Sync milestones data to cache
- Updated `cmd/pm/validate.go`: Check for broken milestone references
- Unit tests in `internal/ticket/ticket_test.go`
- Integration tests in `integration_test.go`

