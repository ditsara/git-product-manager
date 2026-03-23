---
id: GPM-55
title: "Implement pm list --milestone filtering"
type: task
status: backlog
priority: medium
points: 2

parent: GPM-14
depends_on: [GPM-52, GPM-54]
blocks: []
related: []

labels: [milestone, filtering]
assignee: ""
created_at: "2026-02-08T15:04:52Z"
updated_at: "2026-02-08T15:04:52Z"
---

# Description

[Claude Haiku 4.5]

**Task:** Add filtering capability to `pm list` to display only tickets assigned to a specific milestone.

## Overview

This task extends the existing `pm list` command with a new `--milestone <id>` flag, enabling users to view all tickets associated with a particular milestone. This is essential for project managers to track milestone-specific work.

## Implementation Steps

- [ ] Add `--milestone <id>` flag to `pm list` command
  - Accept single milestone ID: `pm list --milestone v1-0-release`
  - Query SQLite cache for all tickets where milestones field contains the ID
- [ ] Implement milestone ID validation in flag parsing:
  - Warn if milestone doesn't exist: "Warning: Milestone 'invalid-id' not found"
  - Still display results if milestone exists but has no tickets
- [ ] Update filtering logic in `cmd/pm/list.go`:
  - Parse milestones field from each ticket (comma-separated)
  - Include ticket if its milestones field contains the target milestone ID
  - Combine with existing filters (--status, --assignee, --label, --parent) using AND logic
  - Example: `pm list --milestone v1-0 --status todo` shows only TODO tickets in v1-0 milestone
- [ ] Add completion support:
  - Update shell completion to suggest milestone IDs for --milestone flag
  - Scan `.pm/milestones/` directory during completion
  - Same case-insensitive matching as ticket IDs
- [ ] Test scenarios:
  - Single milestone filter: `pm list --milestone v1-0-release` → shows all tickets
  - Combined filters: `pm list --milestone v1-0 --status done` → shows only done tickets
  - Non-existent milestone: `pm list --milestone fake-id` → warns but shows empty list (no error)
  - Multiple milestone assignment: Ticket assigned to ["v1-0", "sprint-3"] appears in both `pm list --milestone v1-0` and `pm list --milestone sprint-3`

## Acceptance Criteria

- [ ] `pm list --milestone v1-0` displays only tickets in that milestone
- [ ] Can combine --milestone with --status, --assignee, --label filters
- [ ] Warning displayed if milestone doesn't exist
- [ ] Shell completion suggests milestone IDs for --milestone flag
- [ ] Empty list displayed if milestone exists but has no tickets
- [ ] Ticket with multiple milestones appears in results for each
- [ ] Integration test: create milestone → assign multiple tickets → verify list filtering

## Code Output

- Updated `cmd/pm/list.go`: Add `--milestone` flag and filtering logic
- Updated `internal/cache/query.go`: Add `MilestoneFilter string` field to `ListOptions` struct; filter using `milestones LIKE '%' || ? || '%'` (handles comma-separated storage from GPM-54)
- Updated `cmd/pm/completion.go`: Add milestone completion function (scan `.pm/milestones/` for IDs)
- Unit tests in `cmd/pm/list_test.go` (if exists, or in `integration_test.go`)
- Integration tests in `integration_test.go`

## Dev Readiness Notes

- Added GPM-52 to `depends_on`: the milestones infrastructure (directory, validation, `internal/milestone` package) must exist before this can filter on milestone IDs.
- The `ListOptions` struct in `internal/cache/query.go` needs a new `MilestoneFilter string` field — this is the key cache-layer change. The LIKE query works with the comma-separated TEXT storage defined in GPM-54.
- Shell completion for `--milestone` should scan `.pm/milestones/` at completion time (not the cache) for freshness — consistent with how `--parent` completion works for ticket IDs.
- The `--milestone` warning for non-existent milestones should check the filesystem, not the cache.

