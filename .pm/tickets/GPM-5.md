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
related: [GPM-4, GPM-70]
labels: [validation, data-quality]
assignee: ""
created_at: "2026-02-03T03:26:08Z"
updated_at: "2026-02-03T03:26:08Z"
---

# Description

Canonical spec for all ticket validation. **GPM-70** (milestone reference checking) is a sub-task scoped to one check within Layer 2.

**Key Principle:** Strict validation with explicit opt-in fixing. NO auto-fix on save by default.

## Two-Layer Validation Strategy

### Layer 1 — Write-time guardrails (inline in `new`, `edit`, `move`)

Runs before every file save and blocks the write on failure. Scoped to the ticket being written.

| Check | Status |
|-------|--------|
| Required fields, ID format, timestamp format | `ticket.Validate()` exists — **not yet called on writes** |
| Enum validation (`type`, `status`, `priority`) | Extend `ticket.Validate()` with config lookup |
| Reference integrity: `parent`, `depends_on`, `milestones` for this ticket | New `ticket.ValidateRefs()` — filesystem scan, not cache |
| YAML syntax | Already enforced by parser on load |

### Layer 2 — `pm validate` (batch, cross-ticket, CI-friendly)

Scans all tickets in `.pm/tickets/`, exits with code 1 on any error. Catches drift from manual edits that bypass Layer 1.

| Check | Notes |
|-------|-------|
| All Layer 1 checks, applied retroactively | Catches manually-edited files |
| Orphaned `milestones:` refs (→ `.pm/milestones/`) | GPM-70 scope; scan filesystem, not cache |
| Orphaned `parent` / `depends_on` refs (→ deleted tickets) | Cross-ticket, batch-only |
| Bidirectional consistency (`blocks` ↔ `depends_on`) | Needs full ticket graph |
| `pm validate --fix` interactive repair | Phase 3 below |

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

### Phase 1: Wire existing validation into writes (quick win)
- [ ] Call `ticket.Validate()` in `pm new`, `pm edit`, `pm move` — it already exists, just unwired
- [ ] Add enum checks to `ticket.Validate()`: `type` vs config, `status` vs workflow.yaml, `priority`
- [ ] Add `ticket.ValidateRefs(fs)` for per-write reference integrity (`parent`, `depends_on`, `milestones`)
- [ ] Add type validation: array fields must be arrays, `points` must be numeric

### Phase 2: `pm validate` command
- [ ] Create `cmd/pm/validate.go`
- [ ] Create `internal/ticket/validator.go` to house batch validation logic
- [ ] Scan all tickets in `.pm/tickets/`
- [ ] Apply all Layer 1 checks retroactively
- [ ] Check orphaned `milestones:` refs against `.pm/milestones/` filesystem (GPM-70)
- [ ] Check orphaned `parent` / `depends_on` refs across full ticket graph
- [ ] Check bidirectional consistency (`blocks` ↔ `depends_on`)
- [ ] Report errors with file path and field name; exit code 1 on any error
- [ ] Colorize output (green ✓, red ✗)

### Phase 3: `pm validate --fix`
- [ ] Add `--fix` flag to validate command
- [ ] For each fixable error, show before/after diff
- [ ] Prompt user to confirm each fix (interactive mode)
- [ ] Apply fixes and update `updated_at` timestamp

### Phase 4: Error message quality
- [ ] Type errors: `Field 'labels' must be an array, got string "bug,critical". Change to: labels: [bug, critical]`
- [ ] Enum errors: `Invalid status 'completed'. Valid states: [backlog, todo, in-progress, done]. Did you mean 'done'?`
- [ ] Format errors: `Invalid timestamp '2026-01-31'. Must be ISO8601 with timezone: '2026-01-31T09:00:00Z'`
- [ ] Ref errors: `GPM-5 references unknown milestone "old-sprint". Run: pm edit GPM-5 --field milestones= (to clear)`
- [ ] Include file path and field name in all errors

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



