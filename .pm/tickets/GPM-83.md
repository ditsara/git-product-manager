---
id: GPM-83
title: "Add pm ai init command to bootstrap LLM agent files"
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

Add `pm ai init` — a command that writes a short bootstrap stub file into the correct location for a given LLM tool. The stub tells the LLM that GPM is installed and how to get context, keeping nothing static that could go stale.

## Design

**Bootstrap stub content** (same for all tools):

```markdown
# GPM — Git Product Manager

This project uses GPM for ticket management. Do not create or edit ticket files manually.

Run `pm ai guide --help` to see available guidance sections.
Run `pm ai guide workflow` before starting any task.
Run `pm list` to see open tickets. Run `pm show <id>` to read a spec.
```

**File targets by tool:**

| Tool | File |
|------|------|
| `claude` | `CLAUDE.md` |
| `copilot` | `.github/copilot-instructions.md` |
| `cursor` | `.cursor/rules/gpm.mdc` |
| `aider` | `CONVENTIONS.md` |

**Command flags:**

```
pm ai init [--for <tool>] [--force]
```

- `--for`: one of `claude`, `copilot`, `cursor`, `aider`, or `all` (default: `all`)
- `--force`: overwrite existing file(s); without it, skip any file that already exists and print a notice

## Implementation

**`cmd/pm/ai_init.go`**

- Define `aiInitCmd` registered under `aiCmd`
- Stub content defined as a `const` in the same file
- For each targeted tool, check if file exists:
  - Exists + no `--force` → print `⚠ Skipped <file> (already exists — use --force to overwrite)`
  - Exists + `--force` → overwrite, print `✓ Overwrote <file>`
  - Does not exist → write, print `✓ Created <file>`
- Create parent directories as needed (e.g. `.github/`, `.cursor/rules/`)

## Acceptance Criteria

- [ ] `pm ai init` writes stub to all tool targets that don't yet exist
- [ ] `pm ai init --for claude` writes only `CLAUDE.md`
- [ ] Existing files are skipped with a notice (no silent overwrite)
- [ ] `pm ai init --force` overwrites existing files
- [ ] Parent directories are created if absent
- [ ] Stub content is a `const` in the binary — no external template files
- [ ] `make test` passes
