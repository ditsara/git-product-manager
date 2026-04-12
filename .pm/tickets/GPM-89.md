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
parent: GPM-48
points: 0
priority: medium
related: []
status: backlog
title: Add color support to pm tree
type: task
updated_at: "2026-04-12T09:34:54Z"
---

## Overview

Add color support to the pm tree command to enhance visual differentiation between ticket types and statuses.

## Features to Implement

- Different colors for different ticket types (epic, story, task, bug)
- Highlight active tickets (todo, in-progress) vs completed tickets (done)
- Optional flag to disable colors (e.g., --no-color)
- ANSI color codes for terminal compatibility

## Implementation Details

- Use existing color utilities if available in codebase
- Ensure compatibility with all terminal types
- Test with various color schemes
- Make colors configurable/optional

## Example Output

- Epic: Bold blue
- Story: Blue
- Task: Green
- Bug: Red
- Done/Completed: Dim/gray

## References

- Parent: GPM-48 (pm tree command)
- Optional feature marked in original spec
- Enhances UX of hierarchy visualization