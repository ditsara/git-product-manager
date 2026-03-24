---
assignee: ""
blocks: []
created_at: "2026-02-04T05:30:49Z"
depends_on: []
id: GPM-28
labels:
    - dx
    - cli
    - llm
parent: ""
points: 5
priority: medium
related: []
status: done
title: Embed Development Workflow Guidance into CLI
type: epic
updated_at: "2026-03-24T00:44:46Z"
---


# Description

[Claude Sonnet 4.6]

LLMs working on repos that *use* GPM have no access to this repo's `AGENTS.md`.
This epic makes GPM self-documenting for any AI assistant by embedding workflow
guidance directly in the CLI binary and surfacing it via a `pm guide` command.

## Problem

- `AGENTS.md` lives in the `git-product-manager` repo, not in repos where GPM is *used*
- Without workflow guidance, LLMs make suboptimal decisions about ticket structure, edge cases, and implementation approach
- There is no discoverable, always-current source of guidance for agents working in GPM-managed repos

## Solution

1. **`pm guide [section]`** command — outputs full Markdown guidance, pipeable to any agent context file
2. **`.pm/config/WORKFLOW_GUIDE.md` stub** — created by `pm init`; a lightweight TOC pointing agents to `pm guide`

## Design Decisions

### Output strategy: full by default, subsections available

```bash
pm guide                 # Full output — all sections concatenated as Markdown
pm guide workflow        # Just the workflow section
pm guide schema          # Just the ticket YAML schema
pm guide commands        # Just the commands reference
pm guide principles      # Just the key principles
```

Full output is the default so that `pm guide > CLAUDE.md` (or `AGENTS.md`,
`GEMINI.md`, etc.) works in one command without needing to know section names.

### WORKFLOW_GUIDE.md: stub only (never goes stale)

`pm init` creates `.pm/config/WORKFLOW_GUIDE.md` as a lightweight stub:

```markdown
# GPM Workflow Guide

This project uses Git Product Manager (GPM). For current guidance, run:

  pm guide                 # Full guide (pipe to a file: pm guide > CLAUDE.md)
  pm guide workflow        # Development workflow
  pm guide schema          # Ticket YAML schema
  pm guide commands        # Commands reference
  pm guide principles      # Key principles

To generate a complete guidance file for your LLM:
  pm guide > CLAUDE.md
```

Embedding full content in `WORKFLOW_GUIDE.md` would go stale as the binary is updated. The stub is a permanent pointer to the source of truth.

### Guide content: embedded `.md` files in `internal/guide/`

Content stored as `.md` files embedded into the binary, mirroring the existing `internal/migrations/embed.go` pattern:

```
internal/guide/
  embed.go       # //go:embed *.md; var FS embed.FS
  guide.go       # Section(), Full(), SectionNames() API
  workflow.md
  schema.md
  commands.md
  principles.md
```

This keeps content readable and editable as plain Markdown without touching Go code.

### DevContainer: out of scope

`pm guide` covers universal GPM workflow. Whether an LLM runs inside a devcontainer is repo-specific setup context that belongs in each repo's own `README` or `CLAUDE.md`.

### MCP: complementary, not competing

An MCP server would expose GPM *operations* as tools (create ticket, list tickets, move status). `pm guide` exposes *guidance text* (how to use GPM well). These serve different integration patterns:

- `pm guide` → for CLI-based agents (Copilot CLI, Codex, terminal LLMs)
- Future MCP server → for IDE integrations (VS Code Copilot, Cursor)

No overlap. A future MCP ticket should be created separately if desired.

## Implementation Steps

- [x] Create `internal/guide/` package:
  - `embed.go`: `//go:embed *.md; var FS embed.FS`
  - `guide.go`: `Section(name string) (string, error)`, `Full() string`, `SectionNames() []string`
  - `workflow.md`: 5-step ticket-driven process (idea → ticket → review → implement → done)
  - `schema.md`: ticket YAML front-matter fields with annotated examples
  - `commands.md`: essential `pm` commands cheat sheet
  - `principles.md`: key principles — **must include**: do not commit to git without user approval
- [x] Create `cmd/pm/guide.go`:
  - Import `internal/guide`
  - Full output when no arg; single section when arg given
  - Unknown section → helpful error listing valid names
  - Register `ValidArgs: guide.SectionNames()` for shell completion
- [x] Update `cmd/pm/init.go`:
  - Add `createWorkflowGuide(pmPath)` writing the stub TOC
  - Add `✓ Created workflow guide` to init output
- [x] Update `cmd/pm/completion_helpers.go`:
  - Add `completeGuideSections` func using `guide.SectionNames()`
- [x] Add integration tests:
  - `pm guide` exits 0 and contains all section headers
  - `pm guide schema` contains schema section, not workflow section
  - `pm guide nonexistent` exits non-zero with helpful error
  - `pm init` creates `.pm/config/WORKFLOW_GUIDE.md`

## Acceptance Criteria

- [x] `pm guide` outputs full Markdown covering all sections
- [x] `pm guide workflow` / `schema` / `commands` / `principles` output only that section
- [x] `pm guide` output can be piped: `pm guide > CLAUDE.md`
- [x] `pm guide <invalid>` prints an error listing valid section names
- [x] `principles` section explicitly states: do not commit to git without user approval
- [x] `pm init` creates `.pm/config/WORKFLOW_GUIDE.md` as a stub TOC
- [x] Shell completion suggests section names for `pm guide <tab>`
- [x] Integration tests pass for all above behaviors

## Out of Scope

- DevContainer environment detection
- Per-project customisable guide content
- `--json` flag (no concrete use case yet)
- MCP server (separate ticket if desired)
