---
id: GPM-26
title: "Add visual indicator for tickets with children in pm list"
type: story
status: backlog
priority: medium
points: 2

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""
depends_on: ["GPM-24"]  # Hierarchical filtering must exist first
blocks: []
related: ["GPM-24"]

labels: [ux, visualization, stage-1.6]
assignee: ""
created_at: "2026-02-03T15:22:27Z"
updated_at: "2026-02-03T15:22:27Z"
---

# Description

[Sonnet 4.5]

Add visual indicators to `pm list` output to show which tickets have children, making the hierarchy more discoverable without running additional commands.

## Problem Statement

When viewing `pm list`, there's no way to know which tickets have children without manually running `pm list --parent <id>` for each ticket. This makes it hard to:
- Identify which epics/stories can be drilled into
- Understand the structure of the project at a glance
- Distinguish between leaf tickets (tasks) and parent tickets (epics/stories)

**Current output:**
```bash
$ pm list
ID      TITLE                          TYPE   STATUS      
GPM-1   Stage 2: Collaboration         epic   backlog     
GPM-2   Stage 3: Relationships         epic   backlog     
GPM-6   Fix findTicketByID             task   done        
```

**Desired output:**
```bash
$ pm list
ID      TITLE                          TYPE   STATUS      CHILDREN
GPM-1   Stage 2: Collaboration         epic   backlog     5
GPM-2   Stage 3: Relationships         epic   backlog     0
GPM-6   Fix findTicketByID             task   done        -
```

## User Story

As a project manager, I want to see which tickets have children at a glance, so that I can quickly understand the project structure and know which tickets I can drill into.

## Solution Options

### Option 1: Child Count Column (Recommended)
Add a `CHILDREN` column showing the number of direct children:
- `5` - Has 5 children
- `0` - Has no children (but could have in future)
- `-` - Never checked (or omit for tickets without parent field set)

**Pros:**
- Quantitative information (know how many children before drilling in)
- Clean, professional appearance
- Easy to scan

**Cons:**
- Requires SQL query to count children for each ticket
- Extra column may make wide terminals wider

### Option 2: Visual Marker (Simple)
Add a marker in the title or ID column:
- `GPM-1 [+]` or `📁 GPM-1` - Has children
- `GPM-6` - No children

**Pros:**
- No extra column needed
- Very quick to implement
- Minimal SQL overhead

**Cons:**
- No information about how many children
- May clutter the ID/title display

### Option 3: Type-Based Inference (No Code)
Rely on the `TYPE` column - assume epics/stories have children:
- Users learn: `epic` → has children, `task` → leaf node

**Pros:**
- No implementation needed
- Aligns with typical workflow

**Cons:**
- Not always accurate (empty epics, tasks with subtasks)
- Doesn't show actual structure

## Recommended Approach

**Option 1 (Child Count Column)** with optimizations:
- Add SQL query to count children: `SELECT COUNT(*) FROM tickets WHERE parent = ?`
- Show count only for tickets that have children (others show `-`)
- Make column optional with `--show-children` flag (or always on by default)
- Cache child counts in the SQLite cache for performance

## Implementation Steps

- [ ] Update SQLite cache schema to include `child_count` column
  - Add column: `ALTER TABLE tickets ADD COLUMN child_count INTEGER DEFAULT 0`
  - Or compute dynamically with JOIN query
- [ ] Update `cache.SyncCache()` to compute child counts
  - For each ticket, count children: `SELECT COUNT(*) FROM tickets WHERE parent = ?`
  - Store in cache table
- [ ] Update `cmd/pm/list.go` to include CHILDREN column
  - Add to SELECT query: `SELECT id, title, type, status, child_count`
  - Format output with new column
  - Show `-` for tickets with child_count = 0 (or blank)
- [ ] Update column width calculations for proper alignment
- [ ] Add `--show-children` flag (optional, if making it opt-in)
- [ ] Update help text with new column description

## Display Format

### Default List (Top-Level)
```bash
$ pm list
ID      TITLE                          TYPE   STATUS      CHILDREN
GPM-1   Stage 2: Collaboration         epic   backlog     5
GPM-2   Stage 3: Relationships         epic   backlog     3
GPM-6   Fix findTicketByID             task   done        -
```

### With --parent Filter
```bash
$ pm list --parent GPM-1
ID      TITLE                          TYPE   STATUS      CHILDREN
GPM-19  Implement pm comment           story  backlog     -
GPM-20  Enhanced pm show               story  backlog     -
GPM-21  Implement pm history           story  backlog     -
```

### Deeply Nested
```bash
$ pm list --parent EPIC-1
ID      TITLE                          TYPE   STATUS      CHILDREN
STORY-1  User Authentication           story  backlog     3
STORY-2  API Integration               story  done        0
```

## Performance Considerations

### Option A: Pre-compute in Cache (Recommended)
- Calculate child counts during cache sync
- Store as `child_count` column in tickets table
- Fast lookups, slightly slower sync

### Option B: Dynamic COUNT Query
- Join with subquery: `LEFT JOIN (SELECT parent, COUNT(*) ...) ON ...`
- No extra storage, slightly slower list command

**Recommendation:** Option A (pre-compute) for better `pm list` performance.

## Testing Requirements

### Unit Tests
- [ ] Compute child count correctly (0, 1, many)
- [ ] Handle tickets with no children
- [ ] Handle deeply nested hierarchies

### Integration Tests
- [ ] `pm list` shows child counts
- [ ] Child count updates after adding/removing children
- [ ] Child count correct for nested hierarchies
- [ ] Column alignment works with varying counts (0, 5, 100+)

## Acceptance Criteria

- [ ] `pm list` displays CHILDREN column
- [ ] Shows count of direct children for each ticket
- [ ] Shows `-` (or blank) for tickets with no children
- [ ] Child counts are accurate
- [ ] Column alignment maintained
- [ ] Performance: `pm list` completes in <200ms for 100 tickets
- [ ] All tests pass

## Edge Cases

- **Ticket with 0 children:** Show `-` or blank (not `0`)
- **Ticket with 100+ children:** Ensure column width accommodates
- **Cache out of sync:** Child count may be stale (acceptable, will sync eventually)
- **Orphaned tickets:** If parent doesn't exist, child is counted as parent's child (though parent may not exist)

## Alternative Enhancements

- Color-code tickets with children (e.g., bold or different color)
- Show hierarchy depth: `L1`, `L2`, `L3` (level in tree)
- Tree view mode: `pm list --tree` with ASCII art (separate ticket)
- Expandable rows (interactive CLI - future enhancement)

## Future Work

- **Indented tree view:** Show hierarchy visually with indentation
- **Interactive mode:** Expand/collapse epics in terminal
- **Visual tree:** ASCII art like `tree` command (see GPM-2 for `pm tree`)

## Notes

This enhancement pairs well with GPM-24 (hierarchical filtering). Together they provide:
- **GPM-24:** Navigation (drill down with `--parent`)
- **GPM-26:** Discovery (see what you can drill into)

The child count helps users understand what's "above the fold" before drilling in.