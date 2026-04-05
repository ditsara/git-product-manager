---
assignee: ""
blocks: []
created_at: "2026-04-05T08:45:05Z"
depends_on: []
id: GPM-78
labels: []
parent: ""
points: 0
priority: medium
related: []
status: done
title: Improve UX around users and assignments
type: epic
updated_at: "2026-04-05T10:29:23Z"
---


# Description

Currently, tickets can be assigned to any arbitrary string with no guidance. This epic adds lightweight UX improvements to make assignment more discoverable and consistent for project teams.

## Sub-tickets

| Ticket | Title | Notes |
|--------|-------|-------|
| GPM-29 | Add optional domain suffix for assignees | `pm assign alice` → `alice@company.com` |
| GPM-79 | Add configurable member list with assignee autocomplete | Store team in `project.yaml`; complete on `pm assign <id> <TAB>` |
| GPM-80 | Warn when assigning to user not in member list | Non-fatal warning with instructions to add them |

GPM-80 depends on GPM-79 (needs the member list to check against). GPM-29 can proceed independently.

## Acceptance Criteria

- [ ] GPM-29: `assignee_domain` in project config; domain auto-appended when set
- [ ] GPM-79: `members` list in project config; used for shell completion on `pm assign`
- [ ] GPM-80: Warning printed when assignee is not in `members` list (only when list is non-empty)
- [ ] `make test` passes across all sub-tickets
