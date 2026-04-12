---
assignee: ""
blocks:
    - GPM-91
created_at: "2026-04-12T10:38:38Z"
depends_on: []
id: GPM-92
labels: []
parent: GPM-2
points: 1
priority: medium
related: []
status: done
title: Tighten ID column width to 15 in pm list
type: task
updated_at: "2026-04-12T10:42:43Z"
---


# Description

Reduce the ID column minimum width from 20 → 15 in `pm list`, and implement smart prefix-truncation so the numeric suffix is always preserved.

## Rationale

A realistic ticket ID is at most: 5-char prefix + `-` + 4 digits + ` (+)` = 14 characters. Width 15 fits this with one spare character. Width 20 wastes 5 columns on every row.

## Truncation Behaviour

When the full ID (including ` (+)`) exceeds the column width, truncate only the **prefix** portion using a single `…` (U+2026), keeping `-NNNN` and any ` (+)` suffix intact.

Examples (width = 15):
```
MYPREFIX-1234      →  MYLONGPRE…-1234   (15 chars)
MYPREFIX-1234 (+)  →  MYLON…-1234 (+)  (15 chars)
GPM-42             →  GPM-42            (no truncation needed)
```

## Implementation Notes

- In `cmd/pm/list.go`, change the ID `TableColumn` width from `20` to `15`.
- Add a `truncateID(id string, maxWidth int) string` helper (e.g. in `cmd/pm/common.go` or `table.go`) that:
  1. If `len([]rune(id)) <= maxWidth`, return as-is.
  2. Split on the last `-` to separate prefix from `-NNNN[ (+)]` suffix.
  3. Compute available prefix runes: `maxWidth - len(suffix) - 1` (for the `…`).
  4. Return `prefix[:avail] + "…" + suffix`.
- Apply `truncateID` when building the `displayID` string in the list loop (replacing the current `truncate` call on the ID cell, which uses `...`).
- The existing `truncate` helper (used for TITLE and other columns) is unchanged.

## Acceptance Criteria

- [x] ID column width is 15 in `pm list`
- [x] IDs that fit within 15 chars are displayed unchanged
- [x] IDs exceeding 15 chars truncate the prefix with `…`, preserving `-NNNN` and ` (+)` suffix
- [x] Existing list integration tests pass
- [x] Add a unit test for `truncateID` covering: no truncation, prefix-only truncation, truncation with ` (+)` suffix