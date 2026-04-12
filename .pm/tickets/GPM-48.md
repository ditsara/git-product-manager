---
assignee: ""
blocks: []
created_at: "2026-02-08T14:55:00Z"
depends_on: []
id: GPM-48
labels:
    - relationships
    - cli
    - visualization
parent: GPM-2
points: 3
priority: medium
related: []
status: done
title: Implement pm tree for hierarchy visualization
type: task
updated_at: "2026-04-12T09:32:58Z"
---


# Description

[Claude Sonnet 4.5]

Implement `pm tree` command to visualize ticket hierarchies as ASCII trees, showing parent-child relationships recursively.

## Command Specification

### `pm tree <id> [--depth N]`

**Display**: Shows parent-child relationships recursively using ASCII box-drawing characters.

**Arguments**:
- `<id>`: Ticket ID to display as root of tree
- `--depth N`: Maximum depth to display (optional, default: unlimited)

**Example output**:
```
GPM-44: Reliability & Data Integrity
├── GPM-5: Implement Bad YAML validation
├── GPM-9: Auto-recovery on database errors
├── GPM-11: Implement pm repair command
├── GPM-17: Implement cache metadata table
├── GPM-25: Make getEditor cross-platform
└── GPM-27: Auto-update updated_at from git history
```

**With children that have children**:
```
GPM-1: Stage 2: Collaboration & History
├── GPM-3: Implement pm comment command
│   ├── GPM-23: Implement comment editing
│   └── GPM-24: Implement comment deletion
└── GPM-4: Implement pm history command
```

**With depth limit**:
```
pm tree GPM-1 --depth 2
GPM-1: Stage 2: Collaboration & History
├── GPM-3: Implement pm comment command
│   ├── ... (3 more children)
└── GPM-4: Implement pm history command
    └── ... (1 more child)
```

## Implementation Approach

### Algorithm

1. Load root ticket and verify it exists
2. Query all tickets with `parent == <id>` to get direct children
3. For each child, recursively load its children (respecting depth limit)
4. Build tree structure in memory
5. Render using ASCII box-drawing characters

### Rendering

**Tree symbols**:
- `├── ` for non-last children
- `└── ` for last child
- `│  ` for continuation lines (vertical pipe)
- `   ` (3 spaces) for last child continuation

**Truncation**:
- Truncate long titles to 60 chars with `...`
- Show ticket ID and type in parentheses if space allows

### Performance

- Use cache database to query parent-child relationships
- For large trees (500+ tickets), pagination or lazy loading optional
- Load on-demand approach acceptable

## Implementation Steps

- [x] Create `cmd/pm/tree.go`
- [x] Implement tree node structure:
  - [x] Recursive tree building
  - [x] Depth limiting
- [x] Implement rendering:
  - [x] Box-drawing character selection
  - [x] Proper indentation
  - [x] Title truncation
- [x] Add error handling:
  - [x] Ticket doesn't exist
  - [x] Invalid depth parameter
- [ ] Add color support (optional):
  - [ ] Different colors for different ticket types
  - [ ] Highlight completed vs. active tickets
- [x] Add shell completion for ticket IDs

## Examples

```bash
# View epic and all its tasks
pm tree GPM-44

# View with depth limit (prevent huge output)
pm tree GPM-1 --depth 3

# Check a task's children
pm tree GPM-5
GPM-5: Implement Bad YAML validation
(no children)
```

## Edge Cases

- **Ticket doesn't exist**: Error with helpful message
- **No children**: Display just the root ticket with "(no children)" or similar
- **Deep hierarchies**: Handle gracefully with depth limits
- **Circular parent references** (shouldn't exist but handle): Detect and error
- **Very long titles**: Truncate and show count of hidden chars
- **Invalid depth**: Validate it's a positive integer

## Testing

- [x] Unit test: Build tree structure correctly
- [x] Unit test: Respect depth limits
- [x] Unit test: Render box-drawing characters correctly
- [x] Integration test: Display hierarchy for epic with multiple levels
- [x] Integration test: Handle no children case
- [x] Integration test: Truncate long titles
- [ ] Test performance with 1000+ tickets

## Acceptance Criteria

- [x] `pm tree <id>` displays parent-child hierarchy correctly
- [x] Box-drawing characters render properly in all shells
- [x] Depth limiting works (`--depth` flag)
- [x] Tickets without children don't show empty subtrees
- [x] Titles are truncated to maintain readability
- [x] Output is properly indented and formatted
- [x] Error handling covers edge cases
- [x] All tests pass
- [x] Shell completion works for ticket IDs
