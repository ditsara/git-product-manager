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
- [ ] Create milestone templates in `.pm/config/templates/milestone.md`
- [ ] Implement milestone ID validation (kebab-case, alphanumeric + hyphens, valid filename)
- [ ] Create database migration 000005 for milestones table:
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
- [ ] Migration 000005 creates milestones table with all required fields
- [ ] Milestone ID validation rejects invalid characters (spaces, dots, underscores, capital letters)
- [ ] Templates exist in `.pm/config/templates/`
- [ ] Unit tests pass for ID validation
- [ ] Database index on state field exists for performance

## Code Output

- `internal/milestone/milestone.go`: Milestone struct and parsing logic
- `internal/milestone/validator.go`: ID and schema validation
- `migrations/000005_add_milestones.up.sql` and `.down.sql`: Database schema
- Updated `cmd/pm/init.go`: Create `.pm/milestones/` directory
- Updated `cmd/pm/templates/milestone.md`: New template file

