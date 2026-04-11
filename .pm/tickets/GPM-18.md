---
assignee: ""
blocks: []
created_at: "2026-02-03T08:52:05Z"
depends_on: []
id: GPM-18
labels:
    - git
    - collaboration
parent: ""
points: 5
priority: high
related:
    - GPM-1
status: done
title: Implement pm sync for cross-branch ticket synchronization
type: task
updated_at: "2026-04-11T13:04:51Z"
---


# Description

[Sonnet 4.5]

Implement `pm sync` command to synchronize the `.pm/` directory between
branches, ensuring ticket changes and comments are visible across the team even
when working on feature branches.

## Current Problem

All ticket operations modify files in the `.pm/` directory, which are
version-controlled in git. This creates a branch isolation problem:

**Scenario:**
1. Developer creates feature branch `feature/auth`
2. They create tickets or add comments: `pm new "Add OAuth"`
3. These changes only exist in `feature/auth` branch
4. Other developers on `main` or other branches can't see these tickets
5. When the feature branch is merged, ticket changes come with it, but often too late

**Impact:**
- **Delayed visibility:** Team can't see work-in-progress tickets until PR is merged
- **Duplicate work:** Someone else might create the same ticket on another branch
- **Lost comments:** Comments added during feature development aren't visible to reviewers
- **Merge conflicts:** Multiple people editing the same ticket on different branches causes conflicts
- **Broken workflows:** Can't effectively use `pm` for sprint planning when tickets are scattered across branches

## Solution

Add a `pm sync` command that pulls/pushes the `.pm/` directory from a configured sync branch (typically `main` or `develop`).

### Sync Workflow

```bash
# Before creating tickets on a feature branch
pm sync pull    # Get latest tickets from main

# After creating/updating tickets
pm sync push    # Push ticket changes to main (or configured branch)

# Or do both
pm sync         # Equivalent to pull + push
```

### How It Works

The sync operation performs targeted git operations on the `.pm/` directory:

1. **Pull:** Fetch and merge `.pm/` from sync branch into current branch
2. **Push:** Commit current `.pm/` changes and push to sync branch
3. **Conflict handling:** Detect and report merge conflicts in ticket files

### Configuration

Add sync settings to `.pm/config/project.yaml` under the `sync:` key:

```yaml
sync:
  # Branch to sync with (auto-detected from git: main or master)
  # If not set, auto-detects by checking remote HEAD or local branches
  branch: ""  # Empty = auto-detect

  # Auto-sync behavior (default: manual)
  auto_sync:
    pull_on_list: false   # Auto-pull before pm list
    push_on_change: false # Auto-push after pm new/edit/comment

  # Conflict resolution strategy
  conflict_strategy: prompt  # Options: prompt, theirs, ours, manual
```

### Edge Cases

**Case 1: Sync branch doesn't exist**
- Error message: "Sync branch 'main' does not exist. Configure with: pm config set sync.branch <branch>"

**Case 2: Uncommitted changes in .pm/**
- Error message: "You have uncommitted changes in .pm/. Commit or stash them before syncing."
- Suggestion: "Run: git add .pm && git commit -m 'chore(pm): update tickets'"

**Case 3: Merge conflicts during pull**
- Detect conflicts in `.pm/` files
- List conflicted ticket IDs
- Options:
  - Abort sync and restore previous state
  - Accept remote changes (theirs)
  - Keep local changes (ours)
  - Manually resolve (open in editor)

**Case 4: Currently on sync branch**
- `pm sync pull`: No-op with message "Already on sync branch 'main'"
- `pm sync push`: No-op with message "Already on sync branch 'main'"

**Case 5: No remote tracking branch**
- Error: "No remote tracking branch for 'main'. Push it first with: git push -u origin main"

**Case 6: Detached HEAD state**
- Error: "Cannot sync from detached HEAD. Checkout a branch first."

**Case 7: Sync during rebase/merge**
- Error: "Repository is in the middle of a rebase/merge. Complete it before syncing."

**Case 8: Not in a git repository**
- Error: "Not in a git repository. pm sync requires git for version control."
- Suggestion: "Initialize git with: git init"
- Exit gracefully (no crash)

## Implementation Steps

- [ ] **Auto-detect default branch**
  - [ ] Implement `detectDefaultBranch()` function
  - [ ] Check `git symbolic-ref refs/remotes/origin/HEAD` (remote default)
  - [ ] Fallback: check if `main` exists, then `master`
  - [ ] Fallback: use first branch found
  - [ ] Cache result in `sync.branch` after first detection

- [ ] **Configuration**
  - [ ] Extend `.pm/config/project.yaml` schema in `internal/config` with a `sync` section
  - [ ] Implement config loading with auto-detection fallback
  - [ ] Validate sync branch existence after resolution

- [ ] **Git Operations**
  - [ ] Implement `isGitRepo()` - check if current directory is in a git repo
  - [ ] Implement `gitFetchBranch(branch)` - fetch remote branch
  - [ ] Implement `gitMergePath(path, branch)` - merge specific directory from branch
  - [ ] Implement `gitPushPath(path, branch)` - push directory to branch
  - [ ] Implement conflict detection in `.pm/` directory
  - [ ] Handle uncommitted changes check

- [ ] **Command: pm sync pull**
  - [ ] Check if in git repo (error if not)
  - [ ] Resolve sync branch (auto-detect if not configured)
  - [ ] Check current branch (error if on sync branch)
  - [ ] Stash check: error if uncommitted changes in `.pm/`
  - [ ] Fetch sync branch from remote
  - [ ] Merge `.pm/` directory from sync branch
  - [ ] Handle merge conflicts with configured strategy
  - [ ] Update cache after successful pull

- [ ] **Command: pm sync push**
  - [ ] Check if in git repo (error if not)
  - [ ] Resolve sync branch (auto-detect if not configured)
  - [ ] Check current branch (no-op if on sync branch)
  - [ ] Auto-commit changes in `.pm/` with message "chore(pm): sync tickets from {current_branch}"
  - [ ] Checkout sync branch
  - [ ] Cherry-pick or merge `.pm/` changes
  - [ ] Push to remote
  - [ ] Return to original branch
  - [ ] Handle push conflicts (suggest pull first)

- [ ] **Command: pm sync** (combined)
  - [ ] Run pull operation
  - [ ] If successful, run push operation
  - [ ] Atomic rollback if push fails

- [ ] **Configuration commands**
  - [ ] `pm config get sync.branch` - show current sync branch
  - [ ] `pm config set sync.branch <branch>` - change sync branch
  - [ ] `pm config set sync.auto_sync.pull_on_list true` - enable auto-pull

- [ ] **Error handling**
  - [ ] User-friendly error messages for all edge cases
  - [ ] Recovery suggestions in error output
  - [ ] Rollback mechanism for failed operations

## Acceptance Criteria

- [ ] `pm sync pull` fetches and merges `.pm/` from configured sync branch
- [ ] `pm sync push` commits and pushes `.pm/` changes to sync branch
- [ ] `pm sync` performs both pull and push in sequence
- [ ] Auto-detects default branch (main/master) when not configured
- [ ] Configuration in `.pm/config/project.yaml` under `sync` allows overriding the auto-detected sync branch
- [ ] Errors gracefully when not in a git repository
- [ ] Errors gracefully when uncommitted changes exist in `.pm/`
- [ ] Detects and reports merge conflicts with clear resolution options
- [ ] No-ops correctly when already on sync branch
- [ ] Works with remote tracking branches
- [ ] Handles all documented edge cases correctly
- [ ] Cache is updated after sync operations
- [ ] Integration test: sync from feature branch → main → another feature branch

## Future Enhancements

- **Auto-sync mode**: `pm config set sync.auto_sync.push_on_change true` to auto-push after any ticket modification
- **Smart conflict resolution**: Auto-resolve comment conflicts (always merge both)
- **Sync history**: `pm sync log` to show sync operations
- **Selective sync**: `pm sync pull --ticket GPM-123` to sync specific tickets only
- **Branch templates**: Configure different sync branches per ticket type (e.g., bugs → main, features → develop)

## Related Work

- **GPM-1**: Comments will need to sync across branches
- **GPM-7**: Git auto-commit feature can integrate with auto-sync
