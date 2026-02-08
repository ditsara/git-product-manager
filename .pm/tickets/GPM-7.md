---
id: GPM-7
title: "Implement git auto-commit for ticket operations"
type: task
status: canceled
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: [GPM-18]
labels: []
assignee: ""
created_at: 2026-02-03T03:26:09Z
updated_at: 2026-02-08T09:33:00Z
---

# Description

[Sonnet 4.5]

~~Implement git auto-commit functionality for ticket operations to support the "GitOps Workflow" philosophy. Currently, all commands have `// TODO: Add to git staging area` or `// TODO: Auto-commit` comments.~~

## Status: Won't Implement

**Decision:** Auto-commit is not the right approach for this project. Instead, we'll use an explicit sync workflow (see GPM-18).

**Rationale:**
1. **User control:** Auto-commits remove user control over commit messages and timing
2. **Noisy history:** Every ticket operation creates a commit, cluttering git history
3. **Merge conflicts:** Auto-commits on every branch increase conflict likelihood
4. **Lost work:** Failed auto-commits could lose ticket changes
5. **Better alternative:** GPM-18 implements `pm sync` for explicit synchronization to a central branch

## Recommended Workflow (GPM-18)

Instead of auto-commit on every operation:

```bash
# Work on tickets locally (no commits)
pm new "My ticket"
pm edit GPM-1
pm comment GPM-2 -m "Looking good"

# When ready, sync to main branch
pm sync push    # Commits .pm/ changes and pushes to sync branch

# Before starting work, get latest tickets
pm sync pull    # Merges .pm/ changes from sync branch
```

This gives users:
- **Explicit control:** Choose when to share ticket changes
- **Batched commits:** One commit per sync, not per operation
- **Clear workflow:** "sync pull" before work, "sync push" after
- **Conflict awareness:** Sync operation reports merge conflicts
- **Flexibility:** Can still manually commit if desired

## See Also

- **GPM-18:** Implement `pm sync` for cross-branch ticket synchronization (recommended approach)

---

## Original Scope (Archived)

<details>
<summary>Commands that would have auto-committed</summary>

- `pm new` → `chore(pm): Create {id}`
- `pm move` → `chore(pm): Move {id} to {status}`
- `pm edit` → `chore(pm): Edit {id}`
- `pm comment` (Stage 2) → `comment(pm): Add comment to {id}`
- `pm link` (Stage 3) → `chore(pm): Link {id} to {target-id} ({type})`

</details>

<details>
<summary>Implementation approach considered</summary>

**Option 1: Shell out to git CLI** (was recommended)
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

</details>

<details>
<summary>Configuration considered</summary>

Add flag to disable auto-commit:
- `--no-commit` flag on each command
- Global config: `pm config set auto-commit false`

</details>

<details>
<summary>Safety considerations</summary>

- Only commit if `.pm/` is in a git repository
- Don't fail the command if commit fails (warn instead)
- Respect `.gitignore` (don't commit `.cache.db`)
- Check for uncommitted changes before operations (optional)

</details>


