---
assignee: ""
blocks: []
created_at: "2026-04-06T14:22:16Z"
depends_on: []
id: GPM-88
labels: []
parent: ""
points: 0
priority: low
related: []
status: done
title: Add --description flag to pm edit
type: task
updated_at: "2026-04-11T12:04:14Z"
---



# Description

## Motivation

`pm edit` currently only supports `--field` for YAML frontmatter fields, or
opening the ticket in `$EDITOR`. The body/description is not addressable as a
named flag, which makes scripted or automated ticket updates awkward (requires
a custom `EDITOR` shim).

## Proposed Change

Add a `--description` flag (string) to `pm edit`:

```bash
pm edit GPM-50 --description "New body text here"
```

When `--description` is provided, replace the ticket body (everything after the
closing `---` of the frontmatter) with the supplied text and update
`updated_at`, without opening an editor.

Note that `--description ""` should clear the ticket body (replace it with an
empty string), while `--description` alone should trigger an error because an
argument was not provided.

## Implementation Notes

- Parse the ticket file, preserve frontmatter, replace body
- `--description` and `--field` can be combined in one invocation
- Opening an editor must NOT be triggered when either `--description` or `--field` is provided

## Acceptance Criteria

- `pm edit <id> --description <text>` replaces the ticket body and saves the file
- `updated_at` is updated
- `--description` and `--field` can be used together in one invocation
- Editor is not opened when `--description` or `--field` is provided
- Tests cover the new flag
