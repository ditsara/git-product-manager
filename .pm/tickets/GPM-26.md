---
assignee: ""
blocks: []
created_at: "2026-02-03T15:22:27Z"
depends_on:
    - GPM-24
id: GPM-26
labels:
    - ux
    - visualization
    - stage-1.6
parent: GPM-43
points: 2
priority: medium
related:
    - GPM-24
status: done
title: Add visual indicator for tickets with children in pm list
type: story
updated_at: "2026-02-08T10:20:10Z"
---


# Description

[Sonnet 4.5]

Add a simple visual indicator to `pm list` output to show which tickets have children, making the hierarchy more discoverable without running additional commands.

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
ID      TITLE                          TYPE   STATUS      
GPM-1   Stage 2: Collaboration (+)     epic   backlog     
GPM-2   Stage 3: Relationships (+)     epic   backlog     
GPM-6   Fix findTicketByID             task   done        
```

## Solution: Simple (+) Indicator

Add a "(+)" suffix to the title for any ticket that has at least one direct child:
- `GPM-1 (+)` - Has children (can drill down with `--parent GPM-1`)
- `GPM-6` - No children (leaf ticket)

**Benefits:**
- Minimal visual clutter - just a small indicator
- No extra columns needed
- Very quick to implement
- No database schema changes
- Can determine presence on-the-fly with EXISTS query

## Implementation Approach

### Database Query

For each ticket, check if children exist:
```sql
SELECT id, title, type, status,
  CASE WHEN EXISTS(SELECT 1 FROM tickets AS t WHERE t.parent = tickets.id)
    THEN 1 ELSE 0 END AS has_children
FROM tickets
```

### Display Logic

When formatting each row:
```go
var indicator string
if hasChildren {
    indicator = " (+)"
} else {
    indicator = ""
}
fmt.Printf("%-20s %-50s %-10s %-15s\n", 
    id, truncate(title + indicator, 50), ticketType, status)
```

### No Schema Changes Needed

- No migrations required
- No child_count column to maintain
- Computed on-the-fly during query

## Implementation Steps

- [x] Update SQL query in `cmd/pm/list.go` to check for children existence
- [x] Add `has_children` to the SELECT clause using EXISTS subquery
- [x] Modify display logic to append " (+)" to title if has_children
- [x] Ensure truncation includes the "(+)" in character count
- [x] Update help text if needed
- [x] Test with various hierarchy levels

## Acceptance Criteria

- [x] Tickets with direct children show " (+)" after title
- [x] Tickets without children show no indicator
- [x] Indicator is included in title truncation
- [x] Works with all list modes (default, --all, --parent)
- [x] All tests pass
- [x] No performance degradation

## Implementation Notes

**What was done:**
- Simplified approach to use "(+)" indicator instead of child count column
- No schema changes required - uses EXISTS subquery to check for children on-the-fly
- Updated all SQL queries in list.go (4 query variants) to include has_children check
- Modified display logic to append " (+)" when has_children > 0
- Removed CHILDREN column header - now shows indicator inline with title
- All migration files reverted - no database changes needed
- Verified with --parent filter - indicators show correctly
- All 47 tests passing

**Result:**
Tickets now show a simple " (+)" indicator when they have children, making the hierarchy discoverable without extra columns or database changes.

## Edge Cases

- **Deeply nested ticket:** Indicator shows presence of immediate children only (not total descendants)
- **Empty epic:** No "(+)" shown (it has no children, even if it's an epic)
- **Truncated title:** The "(+)" may cause earlier truncation if title is very long
  - Example: "Very long title that would normally truncate" → "Very long title that woul... (+)" (if char limit exceeded)
  - This is acceptable - truncation priority is: keep "(+)" visible if possible

## Advantages Over Count Column

- **Simpler:** No schema changes, no migrations, no synchronization
- **Faster:** Binary check (EXISTS) vs COUNT for performance
- **Cleaner:** Single indicator vs numeric column
- **User-friendly:** Shows "can I drill down?" yes/no without unnecessary numbers
- **Minimal code:** 3-5 lines of logic change vs 30+ lines for child_count feature

## Future Enhancements

- **Tree view:** ASCII art tree display (`pm list --tree`)
- **Expanded view:** Show children inline (with indentation)
- **Color indicator:** Different color/style for parent vs leaf tickets
