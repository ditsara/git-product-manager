# Git Product Manager (GPM)

A Git-based project management system that eliminates the context switch between
coding and project management. All tickets are stored as structured YAML+Markdown
files within the same repository as the code.

**Core principles:**
- **Single Source of Truth** — repository contains both tasks and code
- **GitOps Workflow** — CRUD via CLI, changes committed like code
- **Process as Code** — workflows and labels are version-controlled config
- **Auditability** — git history is the immutable audit trail

This project uses GPM to manage its own development (dogfooding). All tickets are
in `.pm/tickets/`. Use `pm list` to see current work and `pm show TICKET-ID` to
read specifications.

---

## Implementation Reference

### File Architecture

```text
project-root/
├── .pm/
│   ├── tickets/              # Ticket files (PREFIX-N.md) and comment dirs
│   ├── milestones/           # Milestone files (kebab-case-slug.md)
│   ├── config/
│   │   ├── project.yaml      # Project prefix (e.g., prefix: GPM)
│   │   ├── workflow.yaml     # State definitions
│   │   ├── labels.yaml       # Allowed tags
│   │   ├── templates/        # Ticket templates (story, task, bug, epic, milestone)
│   ├── AGENTS.md             # This file
│   ├── .gitignore            # Ignores .cache.db
│   └── .cache.db             # Git-ignored SQLite index
```

### Technology Stack

| Concern          | Library                                          |
|------------------|--------------------------------------------------|
| CLI framework    | `github.com/spf13/cobra` + `github.com/spf13/viper` |
| YAML parsing     | `gopkg.in/yaml.v3` (strict mode)                 |
| Git operations   | Shell out to `git` CLI                           |
| SQLite driver    | `modernc.org/sqlite` (CGo-free)                  |
| ORM              | `github.com/stephenafamo/bob` (Layer 1 only)     |
| Markdown render  | `github.com/charmbracelet/glamour`               |
| DB migrations    | `github.com/golang-migrate/migrate/v4`           |
| Embedded assets  | `embed.FS` (migrations + guide)                  |

### Project Structure

```text
cmd/pm/               # Cobra commands (one file per command)
  main.go
  init.go, new.go, list.go, show.go, move.go, edit.go
  comment.go, history.go, assign.go, ai.go, ai_guide.go, ai_init.go
  link.go, unlink.go, blocked.go, milestone.go
  completion.go, completion_helpers.go, common.go
internal/
  ticket/             # Ticket struct, parser, validator, comment ops
  config/             # workflow.go, labels.go, project config
  cache/              # SQLite ops, Bob ORM queries, migration runner
  migrations/         # Embedded SQL migration files (*.sql)
  guide/              # Embedded guide markdown (*.md)
  milestone/          # Milestone struct, validator
integration_*_test.go # Integration tests (one file per command group)
scripts/
  build.sh            # go build -o bin/pm ./cmd/pm
  test-local.sh       # Full smoke test in sandbox/
Makefile
```

### Build & Development

```bash
make build            # go build -o bin/pm ./cmd/pm
make test             # go test ./...
make test-local       # build + smoke test in sandbox/
make clean            # remove bin/ sandbox/
go install ./cmd/pm   # install to $GOPATH/bin
```

Use `t.TempDir()` for all filesystem-based tests. Integration tests are split
by command group (e.g., `integration_list_test.go`, `integration_ai_test.go`).

### Database Migrations

Migrations are **embedded in the binary** via `internal/migrations/embed.go`
(`//go:embed *.sql`). They are applied automatically on every command invocation
via a lazy migration check.

**Naming:** `{version}_{description}.{up|down}.sql`, zero-padded to 6 digits.

| # | Description                              |
|---|------------------------------------------|
| 1 | Initial schema: `tickets`, `relationships` |
| 2 | Cache metadata table                     |
| 3 | Comments table                           |
| 4 | Relationships table                      |
| 5 | Path column (materialized paths)         |
| 6 | Milestones table                         |
| 7 | `milestones` column on tickets           |
| 8 | `idx_ticket_milestones` index            |

When adding a migration use the next sequential number. Never reuse or skip.

---

## Cache Data Model

The cache is a SQLite database (`.pm/.cache.db`) providing fast ticket queries
without parsing YAML on every command. Ticket files are the source of truth;
the cache is rebuilt automatically when stale.

### Sync Strategy

`ShouldSync()` compares `cache_metadata.last_sync_timestamp` against the
most-recent file mtime in `.pm/tickets/`. If any file is newer, `SyncCache()`
runs a full rebuild inside a transaction.

### Tables

**`tickets`** — primary index of ticket metadata (id, title, type, status,
priority, assignee, parent, created_at, updated_at, body, milestones, path).

**`cache_metadata`** — key/value sync state. Key `last_sync_timestamp` tracks
the last successful sync (ISO8601 UTC). Initialized to epoch to force first
sync.

**`comments`** — index of comment file metadata (ticket_id, author, timestamp,
filepath). Content is NOT cached; `filepath` is used to read on demand.

**`relationships`** — directed graph of ticket relationships
(from_ticket, to_ticket, relationship_type). Types: `depends-on`, `blocks`,
`parent`, `related`.

**`milestones`** — index of milestone metadata (id, title, status, due_date).

### Bob ORM (Layer 1 only)

Use `bob.NewQuery(...)` / `sqlite.Insert(...)` / `sm.Select(...)` etc. for all
DB operations. Do NOT add Layer 2 code generation.

---

## Key Design Decisions

**Ticket IDs — sequential integers, filesystem-first.**
Format: `PREFIX-N` (e.g., `GPM-42`). Generated by scanning `.pm/tickets/` for
the highest existing number. Does NOT rely on the SQLite cache. Gaps in the
sequence are handled gracefully (1,2,4,5 → next is 6).

**Milestone IDs — kebab-case slugs** derived from title (e.g., `v1-0-release`).
Stored in `.pm/milestones/`.

**Ticket IDs are case-insensitive** in all commands. Internally stored as
uppercase; matched case-insensitively on lookup.

**Array fields REPLACE, not append.** When updating `labels`, `depends_on`,
`milestones`, etc. via `pm edit --field`, the new value replaces the entire
array.

**Error messages** must include: ticket ID, field name, expected vs actual
value, and a suggestion.

Example: `Error: Invalid status 'done-ish' for GPM-42. Valid states:
[backlog, todo, in-progress, done, canceled].`

---

## Current Roadmap

**Completed stages:**
- ✅ Stage 1 — Core ticket management (init, new, list, show, edit, move)
- ✅ Stage 1.5 — Sequential IDs, required prefix, config file
- ✅ Stage 1.6 — Lazy migration check, improved help messages
- ✅ Stage 2 — Comments, history, assign, pm show with comments
- ✅ Stage 2+ — pm link, pm unlink, pm blocked, pm assign, bash completion
- ✅ Milestones (GPM-14)
- ✅ Bob ORM migration (GPM-61)
- ✅ pm ai guide (GPM-28, GPM-82)
- ✅ pm ai init + AGENTS.md bootstrap (GPM-83, GPM-87)

**Backlog (run `pm list` for full list):**
- GPM-2 — Stage 3 partial: pm tree, pm search, enhanced list filtering
- GPM-5 — Bad YAML validation guardrails
- GPM-17 — Cache metadata / staleness tracking
- GPM-70 — pm validate command (milestone reference checking)

