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
status: done
title: 'Cache sync: detect and warn on broken relationship symmetry'
type: task
updated_at: "2026-04-12T12:47:42Z"
---


## Overview

Detect broken relationship symmetry during cache sync (when users manually
edit ticket files) and warn users about inconsistencies.

While `pm link` / `pm unlink` maintain symmetry automatically, direct YAML
edits can break the inverse relationship. Two symmetric pairs must be checked:

- `depends_on` ↔ `blocks` (directed: A depends_on B → B blocks A)
- `related` ↔ `related` (bidirectional: A related B → B related A)

Example broken state:

```yaml
# GPM-5.md
depends_on: [GPM-10]
# GPM-10.md
blocks: []   # should contain GPM-5 but does not
```

This creates silent inconsistencies: `pm show GPM-5` reports the dependency but
`pm blocked GPM-10` misses GPM-5.

## Current codebase context (2026-04-12)

- `internal/cache/sync.go` — all sync logic lives here. `scanTicketFiles()`
  already parses DependsOn, Blocks, and Related into `[]relationshipData`.
- The sync pipeline is: parse files → compute paths → syncTickets →
  syncComments → syncRelationships → updateSyncTimestamp. Validation fits
  after `scanTicketFiles` returns, before DB writes.
- `SyncCache` is the single entry point; print warnings to `os.Stderr`.
  Warnings are non-fatal — sync continues regardless.
- No `pm repair` command yet. Users fix problems manually with `pm link`.

## Warning Format

Each warning is two lines: what is wrong, then the exact `pm link` command
to fix it.

```
⚠  GPM-5 depends_on GPM-10, but GPM-10 does not block GPM-5; to fix:
   pm link GPM-10 GPM-5 --type blocks

⚠  GPM-3 related GPM-7, but GPM-7 does not relate back to GPM-3; to fix:
   pm link GPM-7 GPM-3 --type related

⚠  GPM-999 blocks GPM-1000, but GPM-1000 does not depends_on GPM-999; to fix:
   pm link GPM-1000 GPM-999 --type depends-on
```

## Implementation Plan

### New file: internal/cache/validate.go

Keep validation logic separate from sync.go for clarity.

```go
// validateRelationshipSymmetry checks that depends_on/blocks pairs and
// related/related pairs are mirrored. Returns human-readable warning strings.
func validateRelationshipSymmetry(tickets []ticketData) []string
```

Algorithm for `depends_on` ↔ `blocks`:
1. For each ticket A with depends_on=[B,...], expect B.blocks to contain A.
2. For each ticket B with blocks=[A,...], expect A.depends_on to contain B.
3. Report any missing entries.

Algorithm for `related` ↔ `related`:
1. For each ticket A with related=[B,...], expect B.related to contain A.
2. Report any missing entries (only report each pair once).

### Hook into SyncCache (internal/cache/sync.go)

After `scanTicketFiles` returns, call `validateRelationshipSymmetry` and
print each warning to `os.Stderr`.

## Acceptance Criteria

- [ ] `validateRelationshipSymmetry` in `internal/cache/validate.go`
- [ ] Checks `depends_on` ↔ `blocks` symmetry
- [ ] Checks `related` ↔ `related` symmetry
- [ ] Called from `SyncCache` after `scanTicketFiles`; warnings printed to stderr (non-fatal)
- [ ] Warning includes exact `pm link` command to fix
- [ ] Unit tests: missing blocks, orphaned blocks, missing related, clean data
- [ ] Integration test: manually break symmetry, verify warning appears on next `pm list`
