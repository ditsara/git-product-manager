# Git Product Manager (GPM)

A Git-native project management system that stores tickets as structured YAML + Markdown files within your repository, eliminating context-switching between code and project management.

## Philosophy

**Single Source of Truth:** Keep tasks and code together in one repository. No context switching between JIRA tabs and your IDE. Everything you need to understand what to build and why lives alongside the code itself.

**LLM-Native Design:** AI code assistants can read tickets directly from your repository—no API keys, no integrations, no external services. Your LLM has full context of requirements, acceptance criteria, and implementation status without leaving your codebase.

**Familiar Workflow:** Uses the same concepts developers already know from JIRA—epics, stories, tasks, bugs, status workflows, assignees. The only difference: it's all version-controlled files instead of a web UI.

**Additional Benefits:**
- **GitOps Workflow:** All ticket operations are git commits
- **Process as Code:** Workflows and labels are version-controlled
- **Auditability:** Git history is the immutable audit trail

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

### Phase 1 (Core)
- ✅ Initialize `.pm/` directory structure
- ✅ Create tickets from templates (story, task, bug, epic)
- ✅ List and filter tickets
- ✅ Display ticket details with comments
- ✅ Audit trail via git history

### Phase 2 (Management)
- ✅ Update ticket status
- ✅ Edit ticket metadata
- ✅ Add comments (conflict-free)
- ✅ Assign tickets

### Phase 3 (Search)
- ⬜ Full-text search
- ⬜ SQLite cache for performance
- ⬜ Advanced filtering

### Phase 4 (Relationships)
- ✅ Parent-child hierarchies
- ✅ Dependencies (blocks/depends_on)
- ⬜ Visualize ticket trees
- ⬜ Dependency graphs

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please read the specification in AGENTS.md before submitting PRs.
