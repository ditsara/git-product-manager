---
id: GPM-70
title: "pm validate command missing milestone reference checking"
type: bug
status: backlog
priority: medium
points: 0

parent: ""
depends_on: []
blocks: []
related: []

labels: [milestone, validate]
assignee: ""
created_at: "2026-03-24T00:06:45Z"
updated_at: "2026-03-24T00:06:45Z"
---

# Description

[Claude Sonnet 4.6]

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


