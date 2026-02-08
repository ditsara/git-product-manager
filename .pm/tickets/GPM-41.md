---
id: GPM-41
title: "Add 'canceled' state to default workflow template"
type: task
status: done
priority: medium
points: 1

parent: ""
depends_on: []
blocks: []
related: [GPM-7, GPM-39]

labels: [workflow, templates]
assignee: ""
created_at: "2026-02-08T09:32:20Z"
updated_at: "2026-02-08T10:00:00Z"
---

# Description

[Claude Sonnet 4.5]

Add a "canceled" state to the default workflow template to support closing tickets without implementing them. This covers scenarios like: won't-fix, duplicate, obsolete requirements, and rejected ideas.

## Rationale

Currently, the default workflow has 4 states (backlog, todo, in-progress, done), but there's no semantic way to close tickets that won't be implemented. Marking them as "done" is misleading since they weren't completed.

## Benefits

- **Clear semantics**: "done" = implemented, "canceled" = closed-without-implementing
- **State groups alignment**: Both "done" and "canceled" go in the "completed" group (hidden from default list)
- **Decision tracking**: Git history preserves why tickets were rejected
- **Industry standard**: Most issue trackers have this (JIRA: Won't Fix, GitHub: Closed, Linear: Canceled)

## Implementation Tasks

- [x] Update `cmd/pm/templates/workflow.yaml` to include "canceled" state
- [x] Add "canceled" to the "completed" state group
- [x] Add explanatory comments to the template
- [x] Update workflow comments to explain use cases for each state
- [x] Update AGENTS.md if needed (default states section)
- [x] Test that new projects get the updated template
- [x] Verify validation accepts "canceled" as a valid state

## Template Changes

```yaml
# Workflow states define the lifecycle of tickets
# States are labels - no enforced transitions (teams move tickets freely)
states:
  - backlog      # Not yet planned
  - todo         # Ready to start
  - in-progress  # Currently being worked on
  - done         # Completed and implemented
  - canceled     # Closed without implementing (won't-fix, duplicate, obsolete)

# initial_state: The default status for new tickets created via 'pm new'
initial_state: backlog

# state_groups: Semantic groupings for filtering and reporting
# These groups are optional but enable smart default behavior
state_groups:
  # Active work - shown by default in pm list
  active: [todo, in-progress]
  
  # Completed work - hidden by default in pm list
  # Includes both implemented (done) and closed-without-implementing (canceled)
  completed: [done, canceled]
  
  # Everything not yet done
  incomplete: [backlog, todo, in-progress]
```

## Testing

- [x] Run `pm init` in a new directory and verify workflow.yaml has "canceled" state
- [x] Create a ticket and move it to canceled: `pm move TICKET-1 canceled`
- [x] Verify `pm list` hides canceled tickets (after GPM-39 is implemented)
- [x] Verify validation passes with canceled state

## Acceptance Criteria

- [x] Default workflow.yaml template includes "canceled" state
- [x] "canceled" is in the "completed" state_group
- [x] Template has clear comments explaining when to use each state
- [x] New projects initialized with `pm init` get the updated template
- [x] Existing projects continue to work (backward compatible)
- [x] All tests pass