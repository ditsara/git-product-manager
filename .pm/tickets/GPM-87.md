---
id: GPM-87
title: "Fix pm init idempotency: skip existing files, generate missing ones"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-81"  # Parent epic or story
depends_on:
  - GPM-83
blocks: []  # This blocks these tickets
related:
  - GPM-86

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-04-05T09:45:52Z"
updated_at: "2026-04-05T09:45:52Z"
---

# Description

Running `pm init` on an already-initialised project should be safe. Currently
it may overwrite config files. Fix it to skip any file that already exists and
generate only what is missing — including `.pm/AGENTS.md` once GPM-83 adds it.

## Current behaviour

`pm init` calls `createProjectConfig`, `createWorkflowGuide`, etc.
unconditionally. Re-running it risks overwriting user edits to
`project.yaml`, `workflow.yaml`, or label/workflow config.

## Target behaviour

For each file `pm init` would create:

- **File exists** → skip silently (or print `✓ <file> already exists,
  skipping`)
- **File missing** → create as normal, print `✓ Created <file>`

This makes `pm init` safe to re-run and also serves as a repair command:
if `.pm/AGENTS.md` is accidentally deleted, `pm init` regenerates it.

## Scope

All files currently written by `pm init`:
- `.pm/config/project.yaml`
- `.pm/config/workflow.yaml`
- `.pm/config/labels.yaml`
- `.pm/config/WORKFLOW_GUIDE.md`
- `.pm/AGENTS.md` (added by GPM-83)

## Acceptance Criteria

- [ ] Re-running `pm init` on an existing project changes nothing
- [ ] Each skipped file prints a clear message
- [ ] A missing file (e.g. deleted `AGENTS.md`) is regenerated
- [ ] `make test` passes (add integration test for re-init)

# Description

