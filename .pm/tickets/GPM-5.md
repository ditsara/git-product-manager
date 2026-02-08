---
id: GPM-5
title: "Implement Bad YAML validation guardrails"
type: task
status: backlog
priority: medium
points: 5
parent: GPM-44
depends_on: []
blocks: []
related: [GPM-4]
labels: [validation, data-quality]
assignee: ""
created_at: "2026-02-03T03:26:08Z"
updated_at: "2026-02-03T03:26:08Z"
---

# Description

[Sonnet 4.5]

Implement the "Bad YAML" validation guardrails described in AGENTS.md section 3.3. Currently, `internal/ticket.Validate()` exists but is not called in CLI workflows.

**Key Principle:** Strict validation with explicit opt-in fixing. NO auto-fix on save by default.

## Required Validations

### On Write Operations (new, edit, move)

**Required Fields:**
- `id`, `title`, `type`, `status`, `created_at` must be present

**Valid Enums:**
- `type` ∈ {epic, story, task, bug}
- `status` must be in configured states list from workflow.yaml
- `priority` ∈ {low, medium, high, critical}

**Type Validation:**
- Array fields (`labels`, `depends_on`, `blocks`, `related`) must be arrays, not strings
- String fields must not be arrays
- Numeric fields (`points`) must be numbers, not strings

**Format Validation:**
- **Date Format:** ISO8601 with UTC (e.g., `2026-01-31T09:00:00Z`)
- **ID Format:** Must match `{PREFIX}-\d+`

**YAML Syntax:**
- Must parse without errors
- Front matter must be valid YAML

**Reference Integrity:** (Stage 2+)
- All ticket IDs in `parent`, `depends_on`, `blocks`, `related` must exist

## Commands

### `pm validate` - Check all tickets

```bash
# Validate all tickets
pm validate

# Output:
# ✓ GPM-1: Valid
# ✓ GPM-2: Valid
# ✗ GPM-4: Field 'labels' must be an array, got string "bug,usability"
#   Suggestion: Change to: labels: [bug, usability]
# Found 1 error in 3 tickets
```

### `pm validate --fix` - Interactive fix mode

```bash
pm validate --fix

# Output:
# ✓ GPM-1: Valid
# ✓ GPM-2: Valid
# ✗ GPM-4: Field 'labels' must be an array, got string "bug,usability"
#   Fix: labels: [bug, usability]
#   Apply fix? [y/N] y
# ✓ Fixed GPM-4
```

### Validation on write operations

All write commands should validate before saving:
```bash
pm new "My ticket"      # Validates generated template
pm edit GPM-1           # Validates after editor closes
pm move GPM-1 done      # Validates status exists in workflow.yaml
```

If validation fails:
- Show error message with suggestion
- Ask: "Edit again? [y/N]" (for interactive commands)
- Exit with error code 1

## NO Auto-Fix on Save

**Explicitly rejected feature:**
- No `pm config set auto-fix true` option
- Validation should catch problems, not silently fix them
- Forces users to learn correct YAML syntax
- Forces tools to generate correct YAML (see GPM-4)

**Rationale:**
- Predictability: What you write = what gets saved
- Educational: Users learn from validation errors
- Safety: No silent data transformations

## Implementation Steps

### Phase 1: Strict Validation
- [ ] Add type validation to `ticket.Validate()`
  - [ ] Array fields must be arrays
  - [ ] String fields must be strings
  - [ ] Numeric fields must be numbers
- [ ] Call `ticket.Validate()` in `pm new`, `pm edit`, `pm move`
- [ ] Add workflow state validation (check against workflow.yaml)
- [ ] Add ID format validation with regex
- [ ] Add timestamp format validation

### Phase 2: `pm validate` Command
- [ ] Create `cmd/pm/validate.go`
- [ ] Scan all tickets in `.pm/tickets/`
- [ ] Report validation errors with file:line numbers
- [ ] Exit with code 1 if any errors found
- [ ] Colorize output (green ✓, red ✗)

### Phase 3: `pm validate --fix`
- [ ] Add `--fix` flag to validate command
- [ ] For each fixable error, show before/after
- [ ] Prompt user to confirm each fix (interactive mode)
- [ ] Apply fixes and update `updated_at` timestamp
- [ ] Commit fixes to git (if auto-commit enabled)

### Phase 4: Error Messages
- [ ] Create helpful error messages with suggestions
  - Type errors: `Field 'labels' must be an array, got string "bug,critical". Change to: labels: [bug, critical]`
  - Enum errors: `Invalid status 'completed'. Valid states: [backlog, todo, in-progress, done]. Did you mean 'done'?`
  - Format errors: `Invalid timestamp '2026-01-31'. Must be ISO8601 with timezone: '2026-01-31T09:00:00Z'`
- [ ] Include file path and field name in all errors
- [ ] Color-code errors in terminal

## Acceptance Criteria

- [ ] Invalid tickets are rejected with clear error messages
- [ ] Error messages suggest corrections when applicable
- [ ] All write operations validate before saving
- [ ] `pm validate` checks all tickets and reports errors
- [ ] `pm validate --fix` can correct common type issues interactively
- [ ] Type mismatches (string vs array) are detected and reported
- [ ] Tests cover all validation rules
- [ ] Validation does not break existing valid tickets
- [ ] NO silent auto-fixing - all fixes require explicit user action

## Related

GPM-4 fixes `pm edit --field` to generate correct YAML arrays. This ticket ensures all tickets stay valid.



