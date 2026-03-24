---
assignee: ""
blocks:
    - GPM-56
created_at: "2026-02-08T15:04:51Z"
depends_on:
    - GPM-52
    - GPM-54
id: GPM-53
labels:
    - milestone
    - cli
parent: GPM-14
points: 5
priority: medium
related: []
status: done
title: Implement pm milestone create/list/show commands
type: task
updated_at: "2026-03-23T23:47:58Z"
---


# Description

[Claude Haiku 4.5]

**Task:** Implement the core milestone CLI commands for creating, listing, and displaying milestones.

## Overview

This task delivers three essential user-facing commands for milestone management. These commands enable users to create milestones, view all milestones in a table, and inspect individual milestone details.

## Implementation Steps

- [x] Implement `pm milestone create "Title" [--due YYYY-MM-DD] [--description TEXT] [--id CUSTOM-ID]`
  - Generate new milestone ID by slugifying title (kebab-case)
  - If `--id` provided, use that instead (still validated as kebab-case)
  - If generated ID already exists as `.pm/milestones/{id}.md`, error: `"Milestone ID '{id}' already exists. Use --id to specify a unique ID."`
  - Set state to "active"
  - Set created_at to current UTC timestamp
  - Write `.pm/milestones/{id}.md` with YAML front-matter
  - Auto-stage file in git
  - Output: `✓ Created milestone: {id}`
- [x] Implement `pm milestone list [--state active|closed]`
  - Query SQLite cache for all milestones (or generate from filesystem if cache stale)
  - Display table with columns: ID, Title, Due Date, State, Days Until/Days Overdue
  - Color-code: green for active/not-overdue, yellow for approaching due date, red for overdue
  - Sort by due_date ascending (null dates last)
  - Output sample:
    ```
    ID              Title                 Due Date   State   Status
    v1-0-release    Version 1.0 Release   Feb 28     active  32 days
    sprint-3        Sprint 3              Feb 14     active  ⚠ 6 days
    mvp-launch      MVP Launch            Jan 31     active  ⚠ OVERDUE (8 days ago)
    beta-complete   Beta Program Close    Jan 15     closed  -
    ```
- [x] Implement `pm milestone show <milestone-id>`
  - Parse milestone YAML file
  - Display formatted milestone info (title, description, due_date, state, created_at, closed_at)
  - Calculate and display days remaining
  - Show count of associated tickets (from `pm list --milestone <id>` query)
  - Render markdown description with terminal formatting
  - Example output:
    ```
    ID:            v1-0-release
    Title:         Version 1.0 Release
    Description:   First stable release with core features
    State:         active
    Due Date:      Feb 28, 2026
    Days Until:    32
    
    Associated Tickets: 5
    
    (description rendered below)
    ```
- [ ] Implement editor integration (optional, not implemented) for creating milestones interactively
  - If no title provided: `pm milestone create` opens $EDITOR with template
  - Template shows example fields: title, description, due_date
  - Parse editor output to extract values
- [x] Implement `pm milestone close <milestone-id>`: **deferred to GPM-56** — that ticket owns progress tracking and the close workflow
- [x] Update SQLite cache:
  - Sync milestones table from filesystem
  - Keep milestones in cache for fast list/show queries
  - Index on state for filtering

## Acceptance Criteria

- [x] `pm milestone create "v1.0"` creates `.pm/milestones/v1-0.md` with proper YAML
- [x] `pm milestone list` displays all milestones with proper formatting
- [x] `pm milestone list --state closed` filters to closed milestones only
- [x] `pm milestone show v1-0` renders complete milestone details
- [x] Overdue milestones display warning indicator
- [x] `pm milestone close v1-0` updates state and closed_at (**implemented in GPM-56**)
- [x] All commands update SQLite cache
- [x] Integration tests for full workflow: create → list → show → close

## Code Output

- `cmd/pm/milestone.go`: Main milestone commands (create, list, show, close)
- `internal/milestone/operations.go`: Core logic for milestone management
- Unit tests in `internal/milestone/milestone_test.go`
- Integration tests in `integration_test.go`

## Dev Readiness Notes

- Added GPM-54 to `depends_on`: `pm milestone show` needs to query associated ticket count, which requires the milestones field on tickets (GPM-54) and the `ListOptions.MilestoneFilter` query support (GPM-55).
- `pm milestone close` was duplicated here and in GPM-56. Removed from this ticket — GPM-56 owns the close command and progress tracking.
- The `milestone` subcommand should use a Cobra command group. Follow the same pattern as `pm link`/`pm unlink` (separate file, registered in `main.go`).
- `pm milestone show` ticket count can be approximated at this stage by querying the filesystem directly; cache-based filtering is GPM-55's responsibility.

