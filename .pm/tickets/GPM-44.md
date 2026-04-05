---
id: GPM-44
title: "Reliability & Data Integrity"
type: epic
status: backlog
priority: high
points: 0

parent: ""
depends_on: [GPM-10]
blocks: []
related: []

labels: [reliability, validation, cache, cross-platform]
assignee: ""
created_at: "2026-02-08T14:23:48Z"
updated_at: "2026-02-08T14:23:48Z"
---

# Description

[Claude Sonnet 4.5]

Ensure Git Product Manager is resilient, self-healing, and maintains data
integrity across all platforms and user workflows. This epic groups technical
improvements that prevent errors, recover gracefully from failures, and keep
data consistent.

## Goals

1. **Never fail cryptically**: Detect and auto-recover from common errors (corrupted cache, bad YAML)
2. **Always show truth**: Cache stays synchronized with filesystem, timestamps reflect actual modification times
3. **Cross-platform reliability**: Works consistently on Linux, macOS, and Windows
4. **User confidence**: System validates input, prevents bad data, and provides repair tools

## Acceptance Criteria

- [ ] Corrupted or missing cache databases auto-recover without user intervention
- [ ] Invalid YAML is detected and reported with helpful error messages
- [ ] Manual ticket edits are automatically detected and reflected in cache
- [ ] Timestamps are computed from git history for consistency across clones
- [ ] Editor selection works on Windows/PowerShell as well as Unix shells
- [ ] Users have a `pm repair` command for manual troubleshooting

## Validation Strategy

Validation is implemented in two layers. **GPM-5** is the canonical spec; **GPM-70** is a sub-task scoped to milestone reference checking.

### Layer 1 — Write-time guardrails (inline in `new`, `edit`, `move`)

Runs before every file save and blocks the write on failure. Covers the ticket being written only:

| Check | Notes |
|-------|-------|
| Required fields (`id`, `title`, `type`, `status`, timestamps) | `ticket.Validate()` exists but is **unwired** — top priority |
| Enum validation (`type`, `status`, `priority`) | Extend `ticket.Validate()` with config lookup |
| Reference integrity for this ticket's `parent`, `depends_on`, `milestones` | New `ticket.ValidateRefs()` — filesystem scan, not cache |
| YAML syntax | Already enforced by parser on load |

### Layer 2 — `pm validate` (batch, cross-ticket, CI-friendly)

Scans all tickets and exits with code 1 on any error. Catches drift from manual file edits that bypass Layer 1:

| Check | Notes |
|-------|-------|
| All Layer 1 checks, retroactively | Catches manually-edited files |
| Orphaned `milestones:` refs (→ `.pm/milestones/`) | GPM-70 scope |
| Orphaned `parent` / `depends_on` refs (→ deleted tickets) | Cross-ticket, batch-only |
| Bidirectional consistency (`blocks` ↔ `depends_on`) | Needs full graph |
| `pm validate --fix` interactive repair | GPM-5 Phase 3 |

## Child Tickets

### Core Stability
- **GPM-5**: Validation guardrails (write-time + `pm validate` command) — canonical validation spec
  - **GPM-70**: `pm validate` milestone reference checking (sub-task of GPM-5)
- **GPM-9**: Auto-recovery on database errors
- **GPM-11**: `pm repair` command
- **GPM-17**: Cache metadata table for staleness tracking

### Cross-Platform & Consistency
- **GPM-25**: Make getEditor cross-platform compatible
- **GPM-27**: Auto-update updated_at from git history

## Dependencies

All tickets depend on **GPM-10** (database migrations) being complete.

## Implementation Order

1. GPM-17 (cache staleness) — foundation for auto-sync
2. GPM-9 (auto-recovery) — builds on migration system
3. GPM-5 Layer 1 (wire `ticket.Validate()` into writes) — zero new code, immediate win
4. GPM-5 Layer 1 (extend with enum + ref checks) — can be done in parallel with GPM-9
5. GPM-5 Layer 2 (`pm validate` command, including GPM-70 milestone checks)
6. GPM-11 (repair command) — uses validation and sync logic
7. GPM-25 (cross-platform editor) — independent, anytime
8. GPM-27 (git timestamps) — most complex, do last

