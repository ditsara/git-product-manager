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
ID         TITLE                          TYPE   STATUS      
GPM-1 (+)  Stage 2: Collaboration         epic   backlog     
GPM-2 (+)  Stage 3: Relationships         epic   backlog     
GPM-6      Fix findTicketByID             task   done        
```

## Solution: Simple (+) Indicator

Add a "(+)" suffix to the ticket ID for any ticket that has at least one direct child:
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
displayID := id
if hasChildren > 0 {
    displayID = id + " (+)"
}
fmt.Printf("%-20s %-50s %-10s %-15s\n", 
    displayID, truncate(title, 50), ticketType, status)
```

### No Schema Changes Needed

- No migrations required
- No child_count column to maintain
- Computed on-the-fly during query

## Implementation Steps

- [x] Update SQL query in `cmd/pm/list.go` to check for children existence
- [x] Add `has_children` to the SELECT clause using EXISTS subquery
- [x] Modify display logic to append " (+)" to ID if has_children
- [x] Ensure ID column width (20 chars) accommodates the indicator
- [x] Update help text if needed
- [x] Test with various hierarchy levels

## Acceptance Criteria

- [x] Tickets with direct children show " (+)" after ticket ID
- [x] Tickets without children show no indicator
- [x] Indicator fits within the 20-character ID column
- [x] Works with all list modes (default, --all, --parent)
- [x] All tests pass
- [x] No performance degradation

## Implementation Notes

**What was done:**
- Simplified approach to use "(+)" indicator after ticket ID instead of title
- No schema changes required - uses EXISTS subquery to check for children on-the-fly
- Updated all SQL queries in list.go (4 query variants) to include has_children check
- Modified display logic to append " (+)" to ID when has_children > 0
- All migration files reverted - no database changes needed
- Verified with --parent filter - indicators show correctly
- All tests passing

**Result:**
Tickets now show a simple " (+)" indicator when they have children, making the hierarchy discoverable without extra columns or database changes.

## Edge Cases

- **Deeply nested ticket:** Indicator shows presence of immediate children only (not total descendants)
- **Empty epic:** No "(+)" shown (it has no children, even if it's an epic)
- **Long ticket ID:** The "(+)" is included in the 20-character ID column width
  - Most IDs are ~10 chars (e.g., "GPM-43"), so "GPM-43 (+)" = 11 chars fits easily
  - This is acceptable - ID column has plenty of room

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
