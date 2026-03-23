---
assignee: ""
blocks:
    - GPM-55
    - GPM-56
created_at: "2026-02-08T15:04:51Z"
depends_on:
    - GPM-52
id: GPM-54
labels:
    - milestone
    - schema
parent: GPM-14
points: 3
priority: medium
related: []
status: done
title: Add milestone field to ticket schema
type: task
updated_at: "2026-03-23T23:47:58Z"
---


# Description

[Claude Haiku 4.5]

**Task:** Extend the ticket YAML schema to support milestone associations and enable milestone assignment via `pm edit`.

## Overview

This task integrates milestones into the existing ticket system by adding a `milestones` field to ticket metadata. A ticket can belong to multiple milestones, enabling flexible grouping across overlapping milestone scopes.

## Implementation Steps

- [ ] Add `milestones: []` field to ticket YAML schema (array of milestone IDs)
- [ ] Update ticket templates (story, task, bug, epic) to include milestones field
- [ ] Create migration 000007 to alter tickets table (000006 is taken by GPM-52's milestones table):
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

- Updated `internal/ticket/ticket.go`: Add `Milestones []string \`yaml:"milestones,omitempty"\`` field to Ticket struct (plural, array; `omitempty` ensures backward compatibility with existing tickets)
- Updated `internal/ticket/validator.go`: Add milestone reference validation logic
- Updated `cmd/pm/edit.go`: Handle `--field milestones=` assignment (comma-separated list)
- Updated `cmd/pm/templates/*.md`: Add `milestones: []` field to story, task, bug, epic templates
- `internal/migrations/000007_add_milestones_to_tickets.up.sql` and `.down.sql`: `ALTER TABLE tickets ADD COLUMN milestones TEXT`
- Updated `internal/cache/sync.go`: Sync milestones column to cache
- Updated `cmd/pm/validate.go`: Check for broken milestone references
- Unit tests in `internal/ticket/ticket_test.go`
- Integration tests in `integration_test.go`

## Dev Readiness Notes

- Migration number must be **000007** (not "000005b" — golang-migrate requires sequential integers). GPM-52 uses 000006 for the milestones table.
- Field name is `milestones` (plural, array) — consistent with GPM-14's design. Tag it `omitempty` so existing tickets without the field parse cleanly.
- Storing milestones as comma-separated TEXT in SQLite (e.g., `"v1-0-release,sprint-3"`) is acceptable for now; LIKE-based filtering is sufficient until full FTS is needed.
- The `--field milestones=` assignment replaces the whole array (consistent with how other array fields like `labels` work in `pm edit`).
- `pm validate` milestone reference check should scan `.pm/milestones/` filesystem, not the cache, to avoid stale-cache false positives.

