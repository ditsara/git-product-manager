---
assignee: ""
blocks: []
created_at: "2026-02-08T15:04:53Z"
depends_on:
    - GPM-52
    - GPM-53
    - GPM-54
id: GPM-56
labels:
    - milestone
    - progress
parent: GPM-14
points: 5
priority: medium
related: []
status: done
title: Progress tracking and milestone close
type: task
updated_at: "2026-03-23T23:47:58Z"
---


# Description

[Claude Haiku 4.5]

**Task:** Implement progress tracking for milestones and add the `pm milestone close` command.

## Overview

This task adds intelligent progress calculation to milestones and implements the workflow for closing completed milestones. Progress is calculated based on ticket completion state and story points, giving project managers visibility into milestone health.

## Implementation Steps

- [ ] Implement progress calculation logic in `internal/milestone/progress.go`:
  - Count total tickets assigned to milestone
  - Count tickets in "done" state
  - Calculate completion percentage by count: `done_count / total_count * 100`
  - Sum total story points of all tickets in milestone
  - Sum story points of done tickets
  - Calculate completion percentage by points: `done_points / total_points * 100`
  - Calculate days remaining: `due_date - today`
  - Detect overdue: if due_date < today and state != "closed"
  - Return ProgressInfo struct with all metrics

- [ ] Enhance `pm milestone show` to display progress:
  - Show progress bar with ASCII art:
    ```
    Progress:    [████████░░░░░░░░░░] 40% (4/10 tickets)
    By Points:   [██████████░░░░░░░░░░░░] 50% (13/26 points)
    Due:         Feb 28, 2026 (32 days)
    ```
  - Color-code progress bar: green for on-track, yellow for approaching due date, red for overdue
  - Show warning if milestone is overdue:
    ```
    ⚠ OVERDUE: Due Feb 14 (8 days ago)  5 tickets not done
    ```

- [ ] Implement `pm milestone close <milestone-id>`:
  - Validate milestone exists and state is "active"
  - Check if all tickets in milestone are done (optional warning if not)
  - Update milestone state to "closed"
  - Set closed_at timestamp
  - Optional: Require --force flag if any tickets are not done: `pm milestone close v1-0 --force`
  - Auto-commit with message: `chore(pm): Close milestone {id}`
  - Output: `✓ Closed milestone: {id}` with final stats

- [ ] Add `pm milestone list --overdue` flag:
  - Show only active milestones with due_date in the past
  - Sorted by due_date ascending (oldest first)
  - Include "Days Overdue" column
  - Example:
    ```
    ID              Title                Due Date   Days Overdue   Tickets
    mvp-launch      MVP Launch           Jan 31     8             5/7 done
    beta-complete   Beta Program Close   Jan 15     24            12/12 done
    ```

- [ ] Update `pm milestone list` to show progress inline (optional column):
  - Add `--with-progress` flag to show completion percentage in list view
  - Example:
    ```
    ID              Title                Due Date   State   Progress
    v1-0-release    Version 1.0 Release  Feb 28     active  40% (4/10)
    sprint-3        Sprint 3             Feb 14     active  100% (8/8) ✓
    ```

- [ ] Implement warning system for approaching due dates:
  - Warn if milestone due in < 7 days (even if not overdue)
  - Warn if milestone due in < 1 day (critical)
  - Yellow/red color coding in list view

- [ ] Progress metrics are **computed at query time** (not stored in the cache) to avoid stale data:
  - Query tickets for milestone via `ListOptions.MilestoneFilter` (GPM-55)
  - Count total and done tickets in Go; sum points from Ticket structs
  - `days_remaining` and `is_overdue` computed from `due_date` at render time
  - No new migration needed for progress — computed from existing data

- [ ] Test scenarios:
  - Milestone with 0 points: Show progress as "N/A"
  - Milestone with no due date: Don't show "Days Remaining"
  - Milestone with no tickets: Show "0/0 done (0%)"
  - Close milestone: Verify state change and closed_at timestamp

## Acceptance Criteria

- [ ] `pm milestone show v1-0` displays progress bar (by count and points)
- [ ] `pm milestone show` displays warning if overdue
- [ ] `pm milestone close v1-0` updates state to "closed" and sets closed_at
- [ ] `pm milestone close v1-0 --force` closes milestone even if tickets not done
- [ ] `pm milestone list --overdue` shows only overdue milestones
- [ ] `pm milestone list --with-progress` shows completion percentage
- [ ] Days remaining calculated correctly (e.g., "32 days", "OVERDUE (8 days ago)")
- [ ] Milestone with no due_date doesn't show dates in list/show
- [ ] Milestone with 0 points doesn't error (shows "N/A" or "0% by points")
- [ ] Integration test: Full milestone lifecycle (create → assign tickets → track progress → close)

## Code Output

- `internal/milestone/progress.go`: Progress calculation logic (pure computation, no DB writes)
- Updated `cmd/pm/milestone.go`: Enhanced show/list/close commands with progress; `pm milestone close` fully implemented here (stub was deferred from GPM-53)
- Unit tests in `internal/milestone/progress_test.go`
- Integration tests in `integration_test.go`

## Dev Readiness Notes

- Added GPM-52 to `depends_on`: `internal/milestone/progress.go` imports the Milestone struct from GPM-52.
- Progress is computed dynamically from live ticket data (no new cache columns). This avoids a migration and stale-cache bugs. If performance becomes a concern, caching can be added as a follow-up.
- `pm milestone close` is fully owned here. GPM-53 has a placeholder note deferring to this ticket.
- The `--force` flag for closing with incomplete tickets should print a warning listing the unfinished tickets before closing.

