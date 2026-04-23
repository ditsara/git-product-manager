---
assignee: ""
blocks: []
created_at: "2026-02-04T04:34:12Z"
depends_on: []
id: GPM-27
labels:
    - ux
    - cli
parent: GPM-44
points: 1
priority: medium
related: []
status: done
title: Add pm edit --touch to bump updated_at on manual edits
type: task
updated_at: "2026-04-23T16:23:45Z"
---


# Description

When a ticket is edited manually (outside the CLI), the `updated_at` field in its YAML
front-matter stays stale. This means `pm list` sorts it as if unchanged, and the edit
is invisible to sort order.

CLI commands (`pm move`, `pm edit`) already update `updated_at` correctly. This ticket
adds an explicit escape hatch for the manual-edit case.

## Why not auto-detect?

- **mtime** is reset to clone time on `git clone`, so using it as a fallback would stamp
  all tickets with the clone timestamp after every fresh checkout — worse than the status quo.
- **git log** is accurate but requires a subprocess per file and adds meaningful complexity.
- The manual-edit case is rare enough that an explicit opt-in is the right tradeoff.

## Solution

Add a `--touch` flag to `pm edit` that sets `updated_at` to `time.Now()` without opening
the editor or changing any other field.

```bash
# After manually editing a ticket file:
pm edit GPM-42 --touch

# Can be combined with other flags in a single write pass:
pm edit GPM-42 --touch --field priority=high
```

## Implementation

1. **`cmd/pm/edit.go`** — add `--touch` flag:
   - If `--touch` is set (with or without `--field` / `--description`), include
     `updated_at = time.Now().UTC().Format(time.RFC3339)` in the field update pass.
   - If `--touch` is the only flag, apply the single-field update and print
     `✓ Touched updated_at for GPM-42` without opening the editor.

2. **`cmd/pm/ai_init.go`** — update `agentsMDContent` to instruct LLMs:
   - After directly editing a ticket's markdown body, run `pm edit <id> --touch` to
     update the timestamp.

## Acceptance Criteria

- [ ] `pm edit GPM-X --touch` updates `updated_at` to now, prints confirmation, no editor opens
- [ ] `pm edit GPM-X --touch --field priority=high` updates both in a single write pass
- [ ] `pm edit GPM-X` (no flags, opens editor) continues to update `updated_at` as today
- [ ] `agentsMDContent` in `ai_init.go` instructs LLMs to run `pm edit <id> --touch`
      after directly editing ticket content
- [ ] Unit test: `--touch` alone updates `updated_at` and no other fields
- [ ] Unit test: `--touch` combined with `--field` updates both in one pass


