---
assignee: ""
blocks: []
created_at: "2026-03-24T00:06:45Z"
depends_on: []
id: GPM-70
labels:
    - milestone
    - validate
parent: GPM-5
points: 0
priority: medium
related: [GPM-5]
status: backlog
title: pm validate command missing milestone reference checking
type: bug
updated_at: "2026-03-26T02:53:05Z"
---

# Description

**This is a sub-task of GPM-5**, which is the canonical spec for all ticket validation. GPM-70 is scoped specifically to the milestone orphan check in `pm validate` (Layer 2). See GPM-5 for the full two-layer validation strategy and implementation order.

---

`pm validate` does not exist as a command, so milestone reference integrity cannot be checked. GPM-54 originally included a checklist item to implement this, but it was deferred.

## Problem

Tickets can reference milestone IDs in their `milestones:` field that no longer exist (e.g., after a milestone file is manually deleted or renamed). There is no way to detect these orphaned references.

## Expected Behavior

`pm validate` should scan all ticket files and report any `milestones:` entries that reference a milestone ID not found in `.pm/milestones/`. Output should include a suggested fix:

```
Error: GPM-5 references unknown milestone "old-sprint"
  Suggestion: pm edit GPM-5 --field milestones= (to clear) or create the missing milestone
```

## Implementation Notes

- `pm validate` does not exist yet; this requires implementing the command from scratch
- Milestone reference check should scan `.pm/milestones/` from the filesystem (not the cache) to avoid stale-cache false positives
- `internal/ticket/validator.go` should be created to house validation logic
- The command should also check other reference integrity (e.g., `parent`, `depends_on`) — coordinate with or subsume any existing validation plans


