# Git Product Manager (GPM)

A Git-native project management system that stores tickets as structured YAML + Markdown files within your repository, eliminating context-switching between code and project management.

## Philosophy

- **Single Source of Truth:** Tasks and code live in the same repository
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
