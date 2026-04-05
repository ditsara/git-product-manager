---
id: GPM-81
title: "Re-organize AI-related content"
type: epic
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""  # Parent epic (for nested epics)
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-04-05T09:01:14Z"
updated_at: "2026-04-05T09:12:26Z"
---

# Description

Reorganize all LLM/AI-related content under a `pm ai` subcommand. The
current `pm guide` command is replaced by `pm ai guide`. A new `pm ai init`
command writes a short bootstrap stub file to the correct location for the
user's LLM tool.

## Design Principles

- **Single source of truth**: all LLM content lives in the binary
  (`internal/guide/`). No generated files contain full guide text —
  they just tell the LLM how to ask for it.
- **Nothing goes stale**: generated bootstrap files contain only a pointer
  to `pm ai guide`. When the guide is updated, the LLM gets fresh content
  by calling the binary.
- **Short bootstrap**: `pm ai init` writes ~5 lines. The LLM calls
  `pm ai guide <section>` when it needs detail.
- **No silent overwrites**: `pm ai init` skips existing files unless
  `--force` is passed.

## Sub-tickets

| Ticket | Title                                          | Notes                                    |
|--------|------------------------------------------------|------------------------------------------|
| GPM-82 | Reorganize `pm guide` → `pm ai guide`          | Clean rename, no backward compat needed  |
| GPM-83 | Add `pm ai init` bootstrap command             | Writes stub to Claude/Copilot/Cursor/Aider targets |
| GPM-84 | Audit guide content for CLI-first instructions | Ensure LLMs use `pm` commands, not file edits |
| GPM-85 | Add guide content: markdown formatting conventions | 80-col wrap, table alignment rules for ticket descriptions |

GPM-83 and GPM-84 both depend on GPM-82.

## Acceptance Criteria

- [ ] `pm guide` removed; `pm ai guide [section]` works in its place
- [ ] `pm ai init [--for <tool>] [--force]` generates correct bootstrap stub
- [ ] Bootstrap stub is ~5 lines pointing to `pm ai guide --help`
- [ ] Guide content directs LLMs to use `pm` CLI for all ticket operations
- [ ] `make test` passes
