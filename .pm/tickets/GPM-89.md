---
assignee: ""
blocks: []
created_at: "2026-04-12T09:34:39Z"
depends_on: []
id: GPM-89
labels:
    - enhancement
    - cli
    - visualization
    - optional
parent: GPM-2
points: 0
priority: medium
related: []
status: backlog
title: Add consistent color output by status and type across CLI commands
type: task
updated_at: "2026-04-12T10:09:23Z"
---

## Overview

Add ANSI color support across `pm tree`, `pm list`, and `pm search` (and any
future commands) to provide consistent visual differentiation by ticket status
and type. Colors should be defined in a single shared utility so all commands
render identically.

## Features to Implement

- **By status**: color-code tickets based on their status (e.g. in-progress highlighted, done dimmed)
- **By type**: differentiate epics, stories, tasks, and bugs visually
- `--no-color` flag (or respect `NO_COLOR` env var) to disable for non-TTY / scripting use

## Shared Color Utility

Define a central color mapping (e.g. `internal/ui/colors.go`) so that `pm
tree`, `pm search`, and future commands all use the same palette. No
per-command color logic.

## Suggested Palette

- Epic: bold blue
- Story: blue
- Task: green
- Bug: red
- Done / canceled: dim / gray
- In-progress: bold / highlighted

## Acceptance Criteria

- [ ] `pm tree` renders tickets with color by type and status
- [ ] `pm search` renders results with the same color scheme
- [ ] Single shared color utility used by both commands
- [ ] `--no-color` flag and `NO_COLOR` env var both suppress color output
- [ ] All tests pass
