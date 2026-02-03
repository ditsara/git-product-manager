---
id: GPM-7
title: "Implement git auto-commit for ticket operations"
type: task
status: backlog
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: []
assignee: ""
created_at: 2026-02-03T03:26:09Z
updated_at: 2026-02-03T03:26:09Z
---

# Description

[Sonnet 4.5]

Implement git auto-commit functionality for ticket operations to support the "GitOps Workflow" philosophy. Currently, all commands have `// TODO: Add to git staging area` or `// TODO: Auto-commit` comments.

## Scope

Commands that should auto-commit:
- `pm new` → `chore(pm): Create {id}`
- `pm move` → `chore(pm): Move {id} to {status}`
- `pm edit` → `chore(pm): Edit {id}`
- `pm comment` (Stage 2) → `comment(pm): Add comment to {id}`
- `pm link` (Stage 3) → `chore(pm): Link {id} to {target-id} ({type})`

## Implementation Approach

**Option 1: Shell out to git CLI** (Recommended)
```go
func gitCommit(filepath, message string) error {
    exec.Command("git", "add", filepath).Run()
    return exec.Command("git", "commit", "-m", message).Run()
}
```

**Option 2: Use go-git library**
- Pure Go implementation
- No external git dependency
- More complex but more portable

## Configuration

Add flag to disable auto-commit:
- `--no-commit` flag on each command
- Global config: `pm config set auto-commit false`

## Safety Considerations

- Only commit if `.pm/` is in a git repository
- Don't fail the command if commit fails (warn instead)
- Respect `.gitignore` (don't commit `.cache.db`)
- Check for uncommitted changes before operations (optional)

## Implementation Steps

- [ ] Create `internal/git/commit.go` with commit helpers
- [ ] Add `--no-commit` flag to relevant commands
- [ ] Update `pm new` to auto-commit
- [ ] Update `pm move` to auto-commit
- [ ] Update `pm edit` to auto-commit
- [ ] Add configuration option for auto-commit
- [ ] Add tests (mock git operations)
- [ ] Document git integration in README

## Acceptance Criteria

- [ ] Ticket operations create git commits with descriptive messages
- [ ] `--no-commit` flag skips committing
- [ ] Commands don't fail if not in a git repository (warn only)
- [ ] `.cache.db` is never committed
- [ ] Tests verify commit behavior

