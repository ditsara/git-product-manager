---
id: GPM-84
title: "Audit guide content: ensure LLMs are directed to use pm CLI commands"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-81"  # Parent epic or story
depends_on:
  - GPM-82
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-04-05T09:12:26Z"
updated_at: "2026-04-05T09:12:26Z"
---

# Description

Audit the content of all guide sections (`workflow.md`, `schema.md`, `commands.md`, `principles.md` in `internal/guide/`) to ensure every action an LLM might take is expressed as a `pm` CLI command, not as a file operation.

## Motivation

LLMs like GitHub Copilot and Claude Code default to editing files directly when given the opportunity. If the guide says "create a ticket" without specifying *how*, the LLM may create the YAML/Markdown file manually — bypassing ID generation, validation, and cache sync. Every instructional step must be CLI-first.

## Audit Checklist

Read each section and check for any instruction that could be misinterpreted as a file operation. For each, rewrite to use the `pm` command:

| Action | Wrong | Right |
|--------|-------|-------|
| Create ticket | edit `.pm/tickets/X.md` | `pm new "title"` |
| Update status | edit YAML `status:` field | `pm move <id> <state>` |
| Assign ticket | edit YAML `assignee:` field | `pm assign <id> <user>` |
| Add comment | create file in ticket dir | `pm comment <id> "text"` |
| Link tickets | edit YAML `depends_on:` | `pm link <id> <id>` |
| Read ticket | `cat .pm/tickets/X.md` | `pm show <id>` |
| List tickets | `ls .pm/tickets/` | `pm list` |

## Also Check

- The `workflow.md` section should have a clear "do not edit ticket files directly" statement near the top
- The `commands.md` section should be comprehensive — if a command exists, it should be listed
- Cross-check `pm --help` output against `commands.md` to find any missing commands

## Acceptance Criteria

- [ ] No guide section describes a file-system operation where a `pm` command exists
- [ ] `workflow.md` has an explicit "use CLI, not files" rule
- [ ] `commands.md` lists all current `pm` subcommands
- [ ] Changes are minimal — rewrite only where needed, don't pad content

