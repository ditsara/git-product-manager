---
id: GPM-73
title: "Refactor SyncCache: break up 243-line function"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-72"  # Parent epic or story
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-03-26T01:46:01Z"
updated_at: "2026-03-26T01:46:01Z"
---

# Description

`SyncCache` in `internal/cache/sync.go` spans lines 161–403 (~243 lines), making it hard to read and test in isolation.

## What it does today

The function handles the full cache sync in one block:
1. Opens a transaction
2. Clears existing tables (via `clearTable`)
3. Bulk-inserts tickets (Bob INSERT)
4. Bulk-inserts comments (Bob INSERT OR IGNORE)
5. Bulk-inserts relationships (Bob INSERT OR IGNORE)
6. Updates the sync timestamp (Bob INSERT OR REPLACE)

## Goal

Break `SyncCache` into smaller, focused helpers following the pattern already used by `clearTable` and `updateSyncTimestamp`. Suggested split:

- `syncTickets(ctx, tx, tickets)` — wraps the bulk ticket insert (currently L311–331)
- `syncComments(ctx, tx, comments)` — wraps the bulk comment insert (currently L339–358)
- `syncRelationships(ctx, tx, relationships)` — wraps the bulk relationship insert (currently L362–382)

`SyncCache` itself becomes an orchestrator that calls these helpers inside the transaction.

## Acceptance Criteria

- `SyncCache` body is ≤80 lines
- Each extracted helper has a single responsibility and is independently testable
- `make test` passes
