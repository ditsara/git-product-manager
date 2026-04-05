---
assignee: ""
blocks: []
created_at: "2026-03-24T00:41:28Z"
depends_on: []
id: GPM-71
labels: []
parent: ""
points: 0
priority: medium
related: []
status: done
title: Add example common commands to the end of `pm --help`
type: story
updated_at: "2026-04-05T10:59:02Z"
---


# Description

Add a curated "Examples" section to the bottom of `pm --help` output so new
users immediately see the most common commands without reading every subcommand.

## Motivation

The current help output lists subcommands but gives no sense of how they fit
together. A short workflow example dramatically reduces time-to-first-ticket
for a new user.

## Proposed Examples Section

```
Examples:
  # Start a new project
  pm init . --prefix MYPROJECT

  # Create tickets
  pm new "Fix login bug"
  pm new "Auth overhaul" --type epic

  # Browse and triage
  pm list
  pm list --active
  pm list --parent GPM-1
  pm show GPM-42

  # Move work forward
  pm move GPM-42 in-progress
  pm edit GPM-42 --field priority=high
  pm edit GPM-42 --field labels=bug,auth
  pm assign GPM-42 alice

  # Collaborate
  pm comment GPM-42 -m "Blocked on infra"
  pm link GPM-42 --depends-on GPM-10
  pm blocked

  # Milestones
  pm milestone create "v1.0 Release" --due 2026-06-01
  pm milestone list --progress

  # AI / LLM integration
  pm ai init
  pm ai guide
  pm ai guide workflow
```

## Implementation

Set `rootCmd.Example` in `cmd/pm/main.go` — Cobra renders it as part of the
standard help output, correctly indented and labelled. Preferred over a manual
`fmt.Printf` because it integrates with `--help` and shell completion docs.

## Acceptance Criteria

- [x] `pm --help` ends with a clearly labelled examples section
- [x] Examples cover: init, new, list, show, move, edit, assign, comment,
      link, blocked, milestone, and ai subcommands
- [x] `make test` passes
