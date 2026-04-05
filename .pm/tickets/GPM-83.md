---
assignee: ""
blocks: []
created_at: "2026-04-05T09:12:26Z"
depends_on:
    - GPM-82
id: GPM-83
labels: []
parent: GPM-81
points: 0
priority: medium
related: []
status: done
title: Add pm ai init command and generate .pm/AGENTS.md on pm init
type: task
updated_at: "2026-04-05T09:53:54Z"
---


# Description

Two related changes that together give LLMs a discoverable, GPM-owned
entry point without touching any user-managed files.

## Part 1: Generate `.pm/AGENTS.md` on `pm init`

`pm init` already creates `project.yaml`, `workflow.yaml`, etc. Add
`.pm/AGENTS.md` to that set. It is a static file, written once, never
regenerated. Since it lives inside `.pm/`, it is clearly GPM-owned and
users know not to hand-edit it.

**Content** (defined as a `const` in the binary):

```markdown
# GPM — Git Product Manager

This project uses GPM for ticket management.
Do not create or edit ticket files manually — use the `pm` CLI.

Get started:
  pm --help                   # list all commands
  pm ai guide workflow        # read the development workflow
  pm ai guide --help          # see all available guide sections
  pm list                     # show open tickets
  pm show <id>                # read a ticket
```

## Part 2: `pm ai init` appends a pointer to tool config files

`pm ai init` appends a short reference to `.pm/AGENTS.md` into the
user's LLM tool config file(s). It never overwrites — only appends —
so existing content is always preserved. It checks for idempotency
before appending (won't add the line twice).

**Appended text** (same for all tools):

```
# GPM
See .pm/AGENTS.md for project management instructions.
```

**File targets by tool:**

| Tool       | File                              |
|------------|-----------------------------------|
| `claude`   | `CLAUDE.md`                       |
| `copilot`  | `.github/copilot-instructions.md` |
| `cursor`   | `.cursor/rules/gpm.mdc`           |
| `aider`    | `CONVENTIONS.md`                  |

**Command flags:**

```
pm ai init [--for <tool>]
```

- `--for`: one of `claude`, `copilot`, `cursor`, `aider`, or `all`
  (default: `all`)
- No `--force` flag — append is always safe

**Behavior per target:**

- File exists, pointer not yet present → append, print
  `✓ Updated <file>`
- File exists, pointer already present → skip, print
  `✓ <file> already configured`
- File does not exist → create with pointer text, print
  `✓ Created <file>`
- Create parent directories as needed (e.g. `.github/`)

## Implementation

- `cmd/pm/init.go`: call `createAgentsFile(pmPath)` alongside existing
  `createProjectConfig`, `createWorkflowGuide`, etc.
- `cmd/pm/ai_init.go`: define `aiInitCmd` registered under `aiCmd`;
  pointer text and AGENTS.md content as `const` values in the file

## Acceptance Criteria

- [ ] `pm init` generates `.pm/AGENTS.md` with the stub content above
- [ ] `pm ai init` appends the pointer to all tool targets (default)
- [ ] `pm ai init --for claude` appends only to `CLAUDE.md`
- [ ] Append is idempotent — running twice does not duplicate the line
- [ ] Existing file content is never modified, only appended to
- [ ] Parent directories are created if absent
- [ ] All content is `const` in the binary — no external templates
- [ ] `make test` passes
