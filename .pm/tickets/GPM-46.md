---
assignee: ""
blocks: []
created_at: "2026-02-08T14:38:53Z"
depends_on: []
id: GPM-46
labels:
    - cache
    - validation
    - data-integrity
parent: GPM-2
points: 3
priority: medium
related:
    - GPM-45
status: backlog
title: 'Cache sync: detect and warn on broken relationship symmetry'
type: task
updated_at: "2026-04-12T12:11:54Z"
---

## Overview

Detect broken relationship symmetry during cache sync (when users manually
edit ticket files) and warn users about inconsistencies.

While `pm link` / `pm unlink` maintain symmetry automatically, direct YAML
edits can break the inverse relationship:

```yaml
# GPM-5.md
depends_on: [GPM-10]
# GPM-10.md
blocks: []   # should contain GPM-5 but doesn't
```

This creates silent inconsistencies: `pm show GPM-5` reports the dependency but
`pm blocked GPM-10` misses GPM-5.

## Current codebase context (2026-04-12)

- `internal/cache/sync.go` — all sync logic lives here. `scanTicketFiles()`
  already parses DependsOn and Blocks into `[]relationshipData`. This is the
  right place to add symmetry validation.
- The sync pipeline is: parse files → compute paths → syncTickets →
  syncComments → syncRelationships → updateSyncTimestamp. Validation fits
  after parsing, before DB writes.
- `SyncCache` is the single entry point; warnings should be returned from
  it (or printed to stderr) so callers don't need to change their call site.
- There is no `pm repair` command yet; auto-heal is optional scope.
- `cache_metadata` table exists and could store warnings, but logging to
  stderr is simpler for v1.

## Implementation Plan

### New file: internal/cache/validate.go

Keep validation logic separate from sync.go for clarity.

```go
// validateRelationshipSymmetry checks that depends_on / blocks pairs are
// mirrored. Returns a list of human-readable warning strings.
func validateRelationshipSymmetry(tickets []ticketData) []string
```

Algorithm:
1. Build expected-blocks map: for each ticket A with depends_on=[B,...], expect B.blocks to contain A.
2. Build actual-blocks map from ticket data.
3. Report missing entries (A depends on B but B doesn't block A).
4. Report orphaned entries (B blocks A but A doesn't depend on B).

### Hook into SyncCache (internal/cache/sync.go)

After `scanTicketFiles` returns, call `validateRelationshipSymmetry` and
collect warnings. Print each warning to `os.Stderr` prefixed with ⚠.
Warnings are non-fatal — sync continues.

### Optional: pm repair --fix-symmetry (cmd/pm/repair.go)

New command that:
1. Reads all ticket files.
2. Rebuilds blocks arrays from depends_on (source of truth).
3. Writes corrected YAML back.
4. Reports what was fixed.

Not required for first iteration; can be a follow-up.

## Acceptance Criteria

- [ ] `validateRelationshipSymmetry` in internal/cache/validate.go
- [ ] Called from SyncCache after ticket file scan
- [ ] Warnings printed to stderr (non-fatal)
- [ ] Unit tests: missing blocks, orphaned blocks, clean data
- [ ] Integration test: manually break symmetry, verify warning appears
- [ ] Optional: pm repair --fix-symmetry auto-heals