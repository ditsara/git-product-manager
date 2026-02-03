---
id: GPM-12
title: "Fix column alignment in pm list for long titles"
type: bug
status: backlog
priority: low
points: 2

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""
depends_on: []
blocks: []
related: []

labels: [ux, formatting]
assignee: ""
created_at: "2026-02-03T04:02:02Z"
updated_at: "2026-02-03T04:02:02Z"
---

# Description

[Sonnet 4.5]

**Bug:** When ticket titles exceed the column width in `pm list`, they overflow and cause misalignment of subsequent columns (TYPE, STATUS).

## Current Behavior

```
ID                   TITLE                                              TYPE       STATUS         
-----------------------------------------------------------------------------------------------
GPM-9                Auto-recovery on database errors                   task       backlog        
GPM-10               Lazy migration check on every command              task       backlog        
GPM-11               Implement pm repair command                        task       backlog        
GPM-12               Fix column alignment in pm list for long titles    bug        backlog
```

Notice how GPM-12's title is exactly 50 characters (the column width), but if it were longer, it would push the other columns out of alignment.

## Problem

In `cmd/pm/list.go`, the output uses fixed-width formatting:
```go
fmt.Printf("%-20s %-50s %-10s %-15s\n", id, title, ticketType, status)
```

The `%-50s` format specifier doesn't truncate - it only pads. Titles longer than 50 characters overflow.

## Solution Options

### Option 1: Truncate with Ellipsis (Recommended)
```go
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}

// Usage:
fmt.Printf("%-20s %-50s %-10s %-15s\n", 
    id, 
    truncate(title, 50), 
    ticketType, 
    status)
```

Output:
```
GPM-12               Fix column alignment in pm list for long t...     bug        backlog
```

### Option 2: Dynamic Column Width (Terminal-Aware)
Use terminal width to adjust column sizes dynamically:
- Detect terminal width (e.g., 80, 120, 160 columns)
- Allocate space proportionally
- Requires `golang.org/x/term` or similar

### Option 3: Use Table Library
Use a proper table formatting library like `olekukonko/tablewriter` or `charmbracelet/lipgloss`:
- Handles alignment, truncation, borders
- More sophisticated but adds dependency

## Recommendation

Start with **Option 1** (simple truncation) since:
- Zero dependencies
- Predictable output
- Easy to implement
- Solves the immediate problem

Later, could add **Option 2** for better UX in wide terminals.

## Implementation

1. Add `truncate()` helper function to `cmd/pm/list.go`
2. Apply to title column (and any other potentially long fields)
3. Consider adding to other display commands (`show`, `search`)
4. Add tests for edge cases (empty strings, exact width, multibyte chars)

## Acceptance Criteria

- [ ] Titles longer than 50 chars are truncated with "..."
- [ ] Column alignment remains perfect regardless of title length
- [ ] Empty titles don't cause issues
- [ ] Unicode/emoji in titles handled correctly (rune count vs byte count)
- [ ] All columns remain aligned in `pm list` output

## Additional Considerations

- **Multibyte characters**: Use `utf8.RuneCountInString()` not `len()` for accurate character counting
- **Future enhancement**: Add `--full` flag to show untruncated titles
- **Consistency**: Apply same truncation logic to `pm search` if implemented

