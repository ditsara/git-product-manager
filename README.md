# Git Product Manager (GPM)

A Git-native project management system that stores tickets as structured YAML + Markdown files within your repository, eliminating context-switching between code and project management.

This is experimental; do not use for production.

## Philosophy

**Single Source of Truth:** Keep tasks and code together in one repository. No context switching between JIRA tabs and your IDE. Everything you need to understand what to build and why lives alongside the code itself.

**LLM-Native Design:** AI code assistants can read tickets directly from your repository—no API keys, no integrations, no external services. Your LLM has full context of requirements, acceptance criteria, and implementation status without leaving your codebase.

**Familiar Workflow:** Uses the same concepts developers already know from JIRA—epics, stories, tasks, bugs, status workflows, assignees. The only difference: it's all version-controlled files instead of a web UI.

**Additional Benefits:**
- **GitOps Workflow:** All ticket operations are git commits
- **Process as Code:** Workflows and labels are version-controlled
- **Auditability:** Git history is the immutable audit trail; rely on audit and
  team conventions instead of restrictions (e.g., valid state transitions and
  permissions)

## Quick Start

```bash
# Initialize in your project
pm init .

# Create your first ticket
pm new "Add user authentication"

# List tickets
pm list

# View a ticket
pm show PROJ-123

# Update ticket status
pm move PROJ-123 in-progress

# Add a comment
pm comment PROJ-123 -m "Started implementation"
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

Enable tab completion for ticket IDs, commands, and flags. The completion system supports:
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

## Documentation

See [AGENTS.md](AGENTS.md) for the complete technical specification.

## Features

### Stage 1: Core Ticket Management (MVP) ✅
- ✅ Initialize `.pm/` directory structure
- ✅ Create tickets from templates (story, task, bug, epic)
- ✅ List tickets with filtering (status, type)
- ✅ Display ticket details
- ✅ Edit ticket metadata (including `--field` for direct updates)
- ✅ Update ticket status
- ✅ YAML validation and parsing
- ✅ SQLite database for ticket indexing
- ✅ Type-aware field parsing (arrays, integers, enums)

### Stage 1.5: Refinements ✅
- ✅ Sequential ticket IDs (PREFIX-1, PREFIX-2, etc.)
- ✅ Configurable prefix (uppercase)
- ✅ Filesystem-based ID generation
- ✅ Column alignment with truncation

### Stage 1.6: UX Polish & CLI Refinements ✅
- ✅ Help improvements (contextual help for no-arg commands)
- ✅ Lazy cache synchronization with auto-recovery
- ✅ State groups for semantic filtering (--completed, --active, --incomplete)
- ✅ Shell completion for bash, zsh, fish, and PowerShell
- ✅ Visual hierarchy indicators in ticket lists
- ✅ Error message formatting improvements

### Stage 2: Collaboration & History (Planned)
- ⬜ Comment system (conflict-free)
- ⬜ Display comments in `pm show`
- ⬜ State change audit trail (`pm history`)
- ⬜ Assignee shorthand (`pm assign`)

### Stage 3: Relationships & Search (Planned)
- ⬜ Full-text search
- ⬜ Relationship management (`pm link`, `pm unlink`)
- ⬜ Hierarchy visualization (`pm tree`)
- ⬜ Dependency tracking (`pm blocked`)
- ⬜ Advanced filtering

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please read the specification in AGENTS.md before submitting PRs.
