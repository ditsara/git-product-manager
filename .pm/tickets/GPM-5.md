---
id: GPM-5
title: "Implement Bad YAML validation guardrails"
type: task
status: backlog
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: []
assignee: ""
created_at: "2026-02-03T03:26:08Z"
updated_at: "2026-02-03T03:26:08Z"
---

# Description

[Sonnet 4.5]

Implement the "Bad YAML" validation guardrails described in AGENTS.md section 3.3. Currently, `internal/ticket.Validate()` exists but is not called in CLI workflows.

## Required Validations

### On Write Operations (new, edit, move)

**Required Fields:**
- `id`, `title`, `type`, `status`, `created_at` must be present

**Valid Enums:**
- `type` ∈ {epic, story, task, bug}
- `status` must be in configured states list from workflow.yaml
- `priority` ∈ {low, medium, high, critical}

**Format Validation:**
- **Date Format:** ISO8601 with UTC (e.g., `2026-01-31T09:00:00Z`)
- **ID Format:** Must match `{PREFIX}-\d+`

**YAML Syntax:**
- Must parse without errors
- Front matter must be valid YAML

**Reference Integrity:** (Stage 2+)
- All ticket IDs in `parent`, `depends_on`, `blocks`, `related` must exist

## Optional Features (Configurable)

**Auto-Fix Mode** (`pm config set auto-fix true`):
- Fix indentation (2 spaces)
- Trim trailing whitespace
- Ensure newline at EOF
- Log all auto-fixes to stderr

**Strict Mode:**
- Reject tickets with unknown fields
- Require specific metadata fields based on type

## Implementation Steps

- [ ] Call `ticket.Validate()` in `pm new`, `pm edit`, `pm move`
- [ ] Add workflow state validation
- [ ] Add ID format validation with regex
- [ ] Add timestamp format validation
- [ ] Create helpful error messages with suggestions
  - Example: `Invalid status 'completed'. Valid states: [backlog, todo, in-progress, done]. Did you mean 'done'?`
- [ ] Add validation tests for each rule
- [ ] Document validation rules in README

## Acceptance Criteria

- [ ] Invalid tickets are rejected with clear error messages
- [ ] Error messages suggest corrections when applicable
- [ ] All write operations validate before saving
- [ ] Tests cover all validation rules
- [ ] Validation does not break existing valid tickets

