---
assignee: ""
blocks: []
created_at: "2026-04-12T10:12:33Z"
depends_on: []
id: GPM-90
labels: []
parent: ""
points: 0
priority: medium
related: []
status: backlog
title: Support multiple --field flags in pm edit
type: story
updated_at: "2026-04-12T10:13:00Z"
---

## Overview

`pm edit` currently only accepts a single `--field` flag per invocation. Passing multiple `--field` flags silently uses only the last value (standard cobra/pflag behaviour). It would be more ergonomic to allow multiple field updates in one command.

## Desired Behaviour

```bash
pm edit GPM-89 --field parent=GPM-2 --field priority=high
```

Both fields should be applied atomically in a single edit.

## Implementation Notes

- Change the `--field` flag from `StringVar` to `StringArrayVar` (or `StringSliceVar`) so cobra collects all occurrences
- Iterate over all provided `--field` values and apply each in turn before writing the file
- Ensure validation runs on all fields before any write occurs (fail fast, no partial updates)

## Acceptance Criteria

- [ ] Multiple `--field` flags in one invocation all take effect
- [ ] Combining `--field` with `--description` in one call continues to work
- [ ] Invalid field in any position causes the whole command to fail with no changes written
- [ ] All existing tests pass