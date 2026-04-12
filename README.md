# Git Product Manager (GPM)

A Git-native project management system that stores tickets as structured YAML +
Markdown files within your repository, eliminating context-switching between
code and project management.

## Philosophy

**Single Source of Truth:** Keep tasks and code together in one repository. No
context switching between JIRA tabs and your IDE. Everything you need to
understand what to build and why lives alongside the code itself.

**LLM-Native Design:** AI code assistants can read tickets directly from your
repository—no API keys, no integrations, no external services. Your LLM has
full context of requirements, acceptance criteria, and implementation status
without leaving your codebase.

**Familiar Workflow:** Uses the same concepts developers already know from
JIRA—epics, stories, tasks, bugs, status workflows, assignees. The only
difference: it's all version-controlled files instead of a web UI.

**Additional Benefits:**
- **GitOps Workflow:** All ticket operations are git commits
- **Process as Code:** Workflows and labels are version-controlled
- **Auditability:** Git history is the immutable audit trail; rely on audit and
  team conventions instead of restrictions (e.g., valid state transitions and
  permissions)

## Quick Start

```bash
# Initialize in your project
pm init . --prefix MYPROJECT

# Create tickets
pm new "Add user authentication"
pm new "Auth overhaul" --type epic
pm new "Login form" --parent MYPROJECT-2

# Browse and triage
pm list
pm list --active
pm list --parent MYPROJECT-2
pm show MYPROJECT-1

# Move work forward
pm move MYPROJECT-1 in-progress
pm edit MYPROJECT-1 --field priority=high
pm assign MYPROJECT-1 alice

# Collaborate
pm comment MYPROJECT-1 -m "Started implementation"
pm link MYPROJECT-1 --depends-on MYPROJECT-3
pm blocked

# Search
pm search "authentication"

# Milestones
pm milestone create "v1.0 Release" --due 2026-06-01
pm list --milestone v1-0-release

# AI/LLM integration
pm ai init
pm ai guide
```

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/git-product-manager.git
cd git-product-manager

# Build the binary
./scripts/build.sh

# Install to $GOPATH/bin
go install ./cmd/pm

# Or use the local binary
./bin/pm --help
```

### Using Go Install

```bash
go install github.com/yourusername/git-product-manager/cmd/pm@latest
```

## Shell Completion

Enable tab completion for ticket IDs, commands, and flags. The completion
system supports:
- **Ticket IDs**: `pm show GPM-<TAB>` lists all matching tickets (case-insensitive)
- **Commands**: `pm li<TAB>` completes to `pm list`
- **Status values**: `pm move GPM-1 <TAB>` shows valid states from workflow.yaml
- **Flag values**: `pm new --type <TAB>` shows story, task, bug, epic
- **Parent filtering**: `pm list --parent <TAB>` completes ticket IDs

### Bash

Add to `~/.bashrc`:
```bash
eval "$(pm completion bash)"
```

### Zsh

Add to `~/.zshrc`:
```zsh
eval "$(pm completion zsh)"
```

### Fish

Add to `~/.config/fish/config.fish`:
```fish
pm completion fish | source
```

### PowerShell

Add to your PowerShell profile:
```powershell
pm completion powershell | Out-String | Invoke-Expression
```


## Development

### Local Development

```bash
# Build
make build

# Run tests
make test

# Test locally in sandbox
make test-local

# Clean build artifacts
make clean
```

### Development with DevContainer (Isolated Copilot CLI)

This project supports containerized development with GitHub Copilot CLI for
safe AI assistance. The container provides isolation - Copilot can only access
your project directory, not your entire system.

#### Setup (one-time)

```bash
# Build and start the container
./scripts/dev-start.sh

# Authenticate Copilot (first time only)
./scripts/dev-shell.sh
copilot
# Follow interactive authentication flow...
exit
```

#### Workflow

```bash
# 1. Start container (if not running)
./scripts/dev-start.sh

# 2. Edit code with your system editor (Neovim, VS Code, etc.)
nvim cmd/pm/list.go

# 3. Use Copilot from inside container
./scripts/dev-copilot.sh

# 4. Build and test inside container
./scripts/dev-shell.sh
make build
pm list
go test ./...
exit

# 5. Stop container when done (or leave running)
./scripts/dev-stop.sh
```

**Architecture:** Your editor runs on the host, the codebase is mounted into
the container, and Copilot CLI runs inside the container. The `dev-copilot.sh`
wrapper script bridges commands from host to container.

## Documentation

See [AGENTS.md](AGENTS.md) for the complete technical specification.

## Features

- **Ticket management** — Create, edit, move, and view tickets across types: story, task, bug, epic
- **Hierarchies** — Parent/child relationships with `--parent` flag; visualize with `pm tree`
- **Milestones** — Group tickets toward a release goal with progress tracking
- **Relationships & dependencies** — Link tickets with `pm link` (depends-on, blocks, related); surface blockers with `pm blocked`
- **Comments & history** — Conflict-free comment threads; full status-change audit trail via `pm history`
- **Assignees** — Assign tickets with `pm assign`; configurable member list with domain suffix support
- **Full-text search** — Search across ticket IDs, titles, and body content with `pm search`
- **Smart filtering** — Filter by status, type, parent, milestone, assignee, or relationship (--depends-on, --blocks, --related)
- **Shell completion** — Tab completion for ticket IDs, commands, statuses, and flags (bash, zsh, fish, PowerShell)
- **Color output** — Status- and type-aware color coding across all commands; disable with `--no-color`
- **SQLite index** — Fast queries backed by auto-syncing cache; YAML files remain the source of truth
- **AI/LLM integration** — `pm ai guide` provides embedded workflow guidance; `pm ai init` bootstraps your LLM tool config
- **Cross-branch sync** — `pm sync` propagates ticket changes across branches
- **GitOps workflow** — All ticket data is version-controlled; git history is the audit trail
- **Process as code** — Workflows, labels, and members are version-controlled config files

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please read the specification in AGENTS.md before submitting PRs.
