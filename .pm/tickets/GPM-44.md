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

Ensure Git Product Manager is resilient, self-healing, and maintains data integrity across all platforms and user workflows. This epic groups technical improvements that prevent errors, recover gracefully from failures, and keep data consistent.

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

## Child Tickets

### Core Stability
- **GPM-5**: Bad YAML validation guardrails
- **GPM-9**: Auto-recovery on database errors
- **GPM-11**: `pm repair` command
- **GPM-17**: Cache metadata table for staleness tracking

### Cross-Platform & Consistency
- **GPM-25**: Make getEditor cross-platform compatible
- **GPM-27**: Auto-update updated_at from git history

## Dependencies

All tickets depend on **GPM-10** (database migrations) being complete.

## Implementation Order

1. GPM-17 (cache staleness) - Foundation for auto-sync
2. GPM-9 (auto-recovery) - Builds on migration system
3. GPM-5 (validation) - Can be done in parallel
4. GPM-11 (repair command) - Uses validation and sync logic
5. GPM-25 (cross-platform editor) - Independent, can be done anytime
6. GPM-27 (git timestamps) - Most complex, do last

