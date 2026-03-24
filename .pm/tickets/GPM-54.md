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

- [x] Add `milestones: []` field to ticket YAML schema (array of milestone IDs)
- [x] Update ticket templates (story, task, bug, epic) to include milestones field
- [x] Create migration 000007 to alter tickets table (000006 is taken by GPM-52's milestones table):
  ```sql
  ALTER TABLE tickets ADD COLUMN milestones TEXT;  -- JSON array as text
  ```
  - Note: SQLite doesn't support JSON natively in older versions; store as comma-separated IDs or JSON array string
  - Decision: Store as comma-separated IDs for simplicity ("v1-0-release,sprint-3")
- [x] Implement milestone assignment in `pm edit`:
  - `pm edit GPM-1 --field milestones=v1-0-release` (single)
  - `pm edit GPM-1 --field milestones=v1-0-release,sprint-3` (multiple, comma-separated)
  - Parse comma-separated list, validate each milestone ID exists
  - Show error: "Milestone not found: invalid-id"
- [x] Implement milestone removal:
  - `pm edit GPM-1 --field milestones=` (clears all milestones)
- [x] Add validation logic:
  - When setting milestones on a ticket, validate all referenced milestone IDs exist
  - Load milestone IDs from filesystem scan of `.pm/milestones/`
  - Provide helpful error message if milestone doesn't exist
  - Validation runs before file write
- [x] Update SQLite cache:
  - Add milestones column to tickets table
  - Index on milestones for fast filtering: `CREATE INDEX idx_ticket_milestones ON tickets(milestones);` (migration 000008)
  - Sync ticket milestones data from filesystem

## Acceptance Criteria

- [x] Ticket YAML includes `milestones: []` field
- [x] `pm edit GPM-1 --field milestones=v1-0-release` adds milestone to ticket
- [x] `pm edit GPM-1 --field milestones=` clears milestones
- [x] Validation rejects non-existent milestone IDs with helpful error
- [x] SQLite cache stores and indexes milestones for fast queries
- [ ] `pm validate` detects orphaned milestone references — deferred to GPM-70
- [x] Integration test: create milestone → assign to ticket → verify in cache
- [x] Backward compatibility: existing tickets without milestones field work fine

## Code Output

- Updated `internal/ticket/ticket.go`: Add `Milestones []string \`yaml:"milestones,omitempty"\`` field to Ticket struct (plural, array; `omitempty` ensures backward compatibility with existing tickets)
- Updated `cmd/pm/edit.go`: Handle `--field milestones=` assignment (comma-separated list); validate milestone IDs exist before writing
- Updated `cmd/pm/templates/*.md`: Add `milestones: []` field to story, task, bug, epic templates
- `internal/migrations/000007_add_milestones_to_tickets.up.sql` and `.down.sql`: `ALTER TABLE tickets ADD COLUMN milestones TEXT`
- `internal/migrations/000008_add_ticket_milestones_index.up.sql` and `.down.sql`: `CREATE INDEX idx_ticket_milestones ON tickets(milestones)`
- Updated `internal/cache/sync.go`: Sync milestones column to cache
- Unit tests in `internal/ticket/ticket_test.go`
- Integration tests in `integration_test.go`

## Dev Readiness Notes

- Migration number must be **000007** (not "000005b" — golang-migrate requires sequential integers). GPM-52 uses 000006 for the milestones table.
- Field name is `milestones` (plural, array) — consistent with GPM-14's design. Tag it `omitempty` so existing tickets without the field parse cleanly.
- Storing milestones as comma-separated TEXT in SQLite (e.g., `"v1-0-release,sprint-3"`) is acceptable for now; LIKE-based filtering is sufficient until full FTS is needed.
- The `--field milestones=` assignment replaces the whole array (consistent with how other array fields like `labels` work in `pm edit`).
- `pm validate` milestone reference check is tracked in GPM-70.

