---
id: GPM-52
title: "Milestone filesystem & schema setup"
type: task
status: backlog
priority: medium
points: 3

parent: GPM-14
depends_on: []
blocks: [GPM-53, GPM-54]
related: []

labels: [milestone, infrastructure]
assignee: ""
created_at: "2026-02-08T15:04:51Z"
updated_at: "2026-02-08T15:04:51Z"
---

# Description

[Claude Haiku 4.5]

**Task:** Establish the filesystem infrastructure and schema for the milestones feature.

## Overview

This task sets up the foundational structures needed for the milestones system: directory layout, YAML schema, validation logic, and database migration. It's a prerequisite for all subsequent milestone tasks.

## Implementation Steps

- [ ] Create `.pm/milestones/` directory structure (created during init, same as `.pm/tickets/`)
- [ ] Define milestone YAML schema with front-matter fields:
  - `id`: Milestone ID (kebab-case, must be valid filename)
  - `title`: Human-readable milestone name
  - `description`: Optional markdown description
  - `due_date`: ISO8601 date (e.g., "2026-02-28"), optional
  - `state`: Either "active" or "closed"
  - `created_at`: ISO8601 timestamp, auto-set
  - `closed_at`: ISO8601 timestamp, null until milestone is closed
- [ ] Update `cmd/pm/init.go` to create `.pm/milestones/` directory during init (alongside `.pm/tickets/`)
  - The `createDefaultTemplates` function creates 4 templates — update the "✓ Created 4 ticket templates" message to 5 after adding the milestone template
- [ ] Add `cmd/pm/templates/milestone.md` as an embedded template (follows the existing embed pattern: `//go:embed all:templates` in `init.go`); it will be deployed to `.pm/config/templates/milestone.md` by `createDefaultTemplates`
- [ ] Implement milestone ID validation (kebab-case, alphanumeric + hyphens, valid filename)
- [ ] Create database migration 000006 for milestones table (000005 is already taken by `add_path_column`):
  ```sql
  CREATE TABLE milestones (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    due_date TEXT,
    state TEXT NOT NULL,
    created_at TEXT NOT NULL,
    closed_at TEXT
  );
  ```
- [ ] Create index on state field for fast filtering: `CREATE INDEX idx_milestone_state ON milestones(state);`
- [ ] Write unit tests for milestone ID validation (valid: "v1-0", "sprint-1"; invalid: "v1.0", "Sprint 1")
- [ ] Update `.pm/.gitignore` to ensure `.pm/milestones/` is tracked (opposite of `.cache.db`)

## Acceptance Criteria

- [ ] `.pm/milestones/` directory created during `pm init`
- [ ] Milestone YAML schema is documented and validated
- [ ] Migration 000006 creates milestones table with all required fields
- [ ] Milestone ID validation rejects invalid characters (spaces, dots, underscores, capital letters)
- [ ] Templates exist in `.pm/config/templates/`
- [ ] Unit tests pass for ID validation
- [ ] Database index on state field exists for performance

## Code Output

- `internal/milestone/milestone.go`: Milestone struct and parsing logic
- `internal/milestone/validator.go`: ID and schema validation
- `internal/migrations/000006_add_milestones.up.sql` and `.down.sql`: Database schema
- Updated `cmd/pm/init.go`: Create `.pm/milestones/` directory; update template count message to 5
- New embedded template `cmd/pm/templates/milestone.md` (deployed to `.pm/config/templates/milestone.md`)

## Dev Readiness Notes

- Migration 000005 already exists (`add_path_column`); this ticket must use 000006.
- Template file lives in `cmd/pm/templates/` (embedded via `//go:embed all:templates` in `init.go`) and is written to `.pm/config/templates/` at init time by `createDefaultTemplates` — same pattern as existing ticket templates.
- Milestone IDs are kebab-case (e.g., `v1-0-release`), not the `PREFIX-N` pattern used for tickets. Validator is separate from ticket ID validation.

## Questions

**Q: What happens if `.pm/milestones/` does not exist (for example, in a project using an older version of this tool)?**

Lazy-create it. Any `pm milestone` command should call `os.MkdirAll(".pm/milestones/", 0755)` before doing any file I/O — same pattern as how `EnsureCacheReady` in `internal/cache/ensure.go` auto-migrates the DB on every command rather than requiring a manual `pm init` re-run. The directory creation is idempotent, so it's safe to call every time. No error or warning needed — it just works.

**Q: What happens if two titles result in the same kebab-case filename?**

Error with a clear message and offer an escape hatch via `--id`. Before writing the file, check whether `.pm/milestones/{slug}.md` already exists. If it does:

```
Error: Milestone ID 'version-1-0-release' already exists.
Use --id to specify a unique ID: pm milestone create "..." --id my-custom-id
```

The `--id` flag should also accept any user-provided ID (still validated as kebab-case). This makes the collision case explicit and user-resolvable without silent auto-incrementing (which would produce confusing IDs like `version-1-0-release-2`). Add `--id` to the `pm milestone create` command spec in GPM-53.
