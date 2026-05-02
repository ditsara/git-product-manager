---
assignee: ""
blocks: []
created_at: "2026-05-02T06:49:53Z"
depends_on: []
id: GPM-96
labels: []
parent: ""
points: 0
priority: medium
related: []
status: backlog
title: List tickets within a milestone
type: story
updated_at: "2026-05-02T06:52:07Z"
---

# Description

`pm milestone show <id>` should list the milestone's tickets below the progress bar, using the same 4-column table format as `pm list`.

## Intended output

```
ID:          my-milestone
Title:       My Milestone
State:       active
Created:     May 01, 2026

Progress:    [██████░░░░░░░░░░░░░░] 33% (2/6) tickets

ID       TITLE    TYPE    STATUS
TIX-1    etc      story   in-progress
TIX-2    etc      bug     done
```

## Acceptance criteria

- Ticket table appears after the progress section (separated by a blank line) when the milestone has ≥1 ticket.
- Table uses the same 4 columns and `renderTable` styling as `pm list`: **ID, TITLE, TYPE, STATUS**.
- Tickets are sorted by status so that done/canceled tickets appear last (using the workflow's done-state list).
- When a milestone has no tickets, the existing `"0 (none assigned)"` message is kept — no empty table is rendered.
- A `--no-tickets` flag on `pm milestone show` suppresses the ticket table (useful for scripting).

## Implementation notes

- `collectTicketSummaries` already returns the ticket data needed; extend it or use its results to populate table rows.
- Reuse `renderTable`, `TableColumn`, `ui.StatusStyleFunc`, and `ui.TypeStyleFunc` from `list.go` / `common.go`.
