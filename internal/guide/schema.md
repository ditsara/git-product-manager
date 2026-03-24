# GPM Ticket Schema

Tickets are stored as Markdown files with YAML front-matter in `.pm/tickets/`.

## File Format

```
.pm/tickets/PROJ-1.md
```

## YAML Front-Matter Fields

```yaml
---
id: PROJ-1                        # Auto-generated: PREFIX-N (sequential integer)
title: "Implement OAuth2 Login"   # Short, imperative description
type: story                       # epic | story | task | bug
status: backlog                   # Must match a state in .pm/config/workflow.yaml
priority: high                    # low | medium | high | critical
points: 5                         # Story points (0 = unestimated)

parent: PROJ-0                    # Parent ticket ID (optional)
depends_on: [PROJ-2, PROJ-3]      # Tickets that must complete first
blocks: []                        # Tickets this one blocks
related: []                       # Loose associations (duplicates, see-also)

milestones: []                    # Milestone IDs this ticket belongs to
labels: [auth, api]               # Tags from .pm/config/labels.yaml
assignee: alice                   # GitHub username or email (optional)
created_at: "2026-01-31T09:00:00Z"
updated_at: "2026-01-31T09:30:00Z"
---

# Description

Markdown body — problem statement, solution approach, implementation steps,
acceptance criteria, code examples, etc.

## Implementation Steps
- [ ] Step one
- [ ] Step two

## Acceptance Criteria
- [ ] Criterion one
- [ ] Criterion two
```

## Default Workflow States

Defined in `.pm/config/workflow.yaml`:

| State       | Meaning                        |
|-------------|--------------------------------|
| `backlog`   | Not yet scheduled              |
| `todo`      | Scheduled, not started         |
| `in-progress` | Actively being worked on     |
| `done`      | Complete                       |
| `canceled`  | Will not be implemented        |

## ID Format

Ticket IDs use the format `PREFIX-N` where PREFIX is set during `pm init` and N is a
sequential integer. Example: `GPM-42`, `MYAPP-7`.

Milestone IDs use kebab-case slugs derived from the milestone title. Example: `v1-0-release`.
