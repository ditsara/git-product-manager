---
assignee: ""
blocks: []
created_at: "2026-02-03T04:02:02Z"
depends_on: []
id: GPM-12
labels:
    - ux
    - formatting
parent: ""
points: 2
priority: low
related: []
status: done
title: Fix column alignment in pm list for long titles
type: bug
updated_at: "2026-02-03T04:48:36Z"
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

## Solution

Implement truncation with ellipsis to ensure columns always align correctly:

```go
func truncate(s string, maxLen int) string {
    // Use rune count for proper Unicode handling
    runes := []rune(s)
    if len(runes) <= maxLen {
        return s
    }
    return string(runes[:maxLen-3]) + "..."
}

// Usage in list.go:
fmt.Printf("%-20s %-50s %-10s %-15s\n", 
    id, 
    truncate(title, 50), 
    ticketType, 
    status)
```

**Output example:**
```
ID                   TITLE                                              TYPE       STATUS         
-----------------------------------------------------------------------------------------------
GPM-12               Fix column alignment in pm list for long t...     bug        backlog
```

**Why this approach:**
- Zero dependencies
- Predictable output
- Easy to implement and maintain
- Solves the immediate problem
- Works correctly with Unicode (uses rune counting, not byte counting)

## Implementation

1. Add `truncate()` helper function to `cmd/pm/list.go`
2. Apply to title column in the Printf statement
3. Ensure Unicode/emoji handling via `[]rune()` conversion
4. Consider adding to other display commands if they have similar issues

## Edge Cases

- **Empty strings**: Already handled by Printf padding
- **Exact width (50 chars)**: No truncation needed
- **Multibyte characters**: Use `utf8.RuneCountInString()` or `[]rune()` for accurate counting
- **Emoji**: Properly counted as single runes
- **Very short maxLen**: Ensure maxLen > 3 to avoid negative slice index

## Future Enhancement

See GPM-15 for potential upgrade to a table formatting library for more sophisticated output.

## Implementation

1. Add `truncate()` helper function to `cmd/pm/list.go`
2. Apply to title column (and any other potentially long fields)
3. Consider adding to other display commands (`show`, `search`)
4. Add tests for edge cases (empty strings, exact width, multibyte chars)

## Acceptance Criteria

- [x] Titles longer than 50 chars are truncated with "..."
- [x] Column alignment remains perfect regardless of title length
- [x] Empty titles don't cause issues
- [x] Unicode/emoji in titles handled correctly (rune count vs byte count)
- [x] All columns remain aligned in `pm list` output

## Additional Considerations

- **Multibyte characters**: Use `utf8.RuneCountInString()` not `len()` for accurate character counting
- **Future enhancement**: Add `--full` flag to show untruncated titles
- **Consistency**: Apply same truncation logic to `pm search` if implemented

