---
assignee: ""
blocks: []
created_at: "2026-02-03T04:45:48Z"
depends_on:
    - GPM-12
id: GPM-15
labels:
    - enhancement
    - ux
    - polish
parent: ""
points: 3
priority: low
related: []
status: done
title: Use table formatting library for better output
type: task
updated_at: "2026-04-05T10:42:01Z"
---


# Description

[Sonnet 4.5]

**Enhancement:** Replace manual Printf-based table formatting with a proper table library for better visual output and automatic column handling.

**Note:** This should be done after GPM-12 (basic truncation) is implemented. GPM-12 solves the immediate problem; this is a nice-to-have enhancement.

## Current State (After GPM-12)

```
ID                   TITLE                                              TYPE       STATUS         
-----------------------------------------------------------------------------------------------
GPM-12               Fix column alignment in pm list for long t...     bug        backlog
GPM-13               Make ticket IDs case-insensitive in commands      bug        backlog
```

Works, but basic. Manual column width management, simple truncation.

## Proposed Enhancement

Use a table formatting library to get:
- Automatic column width calculation
- Terminal-aware sizing (adjusts to terminal width)
- Better visual styling (borders, colors, alignment)
- Consistent formatting across all list commands
- Less manual string manipulation

## Library Options

### Option A: charmbracelet/lipgloss + bubbles (Recommended)

**Pros:**
- Modern, actively maintained
- Excellent terminal styling support
- Part of Charm ecosystem (glamour already used for markdown)
- Beautiful default styles
- Good Unicode support

**Cons:**
- Adds dependency (but we already use charmbracelet/glamour)
- Might be overkill for simple tables

**Example:**
```go
import "github.com/charmbracelet/lipgloss/table"

t := table.New().
    Headers("ID", "TITLE", "TYPE", "STATUS").
    Rows(
        []string{"GPM-1", "Stage 2: Collaboration", "epic", "backlog"},
        []string{"GPM-2", "Stage 3: Relationships", "epic", "backlog"},
    ).
    StyleFunc(func(row, col int) lipgloss.Style {
        if row == 0 {
            return lipgloss.NewStyle().Bold(true)
        }
        return lipgloss.NewStyle()
    })

fmt.Println(t)
```

### Option B: olekukonko/tablewriter

**Pros:**
- Simple, battle-tested
- Minimal API
- Good for basic tables

**Cons:**
- Less actively maintained
- Simpler styling options
- Another dependency outside Charm ecosystem

### Option C: rodaine/table

**Pros:**
- Very lightweight
- Zero dependencies
- Markdown table output support

**Cons:**
- Limited styling
- No color support

## Recommendation

Use **charmbracelet/lipgloss** table package since:
1. We already depend on Charm ecosystem (glamour for markdown)
2. Consistent theming across the application
3. Modern, well-maintained
4. Terminal-aware and responsive

## Implementation

### Phase 1: Replace `pm list` output
- [ ] Add `github.com/charmbracelet/lipgloss` dependency (if not already present)
- [ ] Replace Printf in `cmd/pm/list.go` with table.New()
- [ ] Configure columns: ID (20 chars), TITLE (auto), TYPE (10), STATUS (15)
- [ ] Add header row styling (bold)
- [ ] Test with various terminal widths

### Phase 2: Apply to other commands
- [ ] Update `pm search` (when implemented)
- [ ] Update `pm milestone list` (when implemented)
- [ ] Consistent table styling across all commands

### Phase 3: Polish
- [ ] Add color coding for status (backlog=gray, in-progress=yellow, done=green)
- [ ] Add alternating row colors for readability
- [ ] Add `--no-color` flag for CI/scripting
- [ ] Responsive column widths based on terminal size

## Before/After

**Before (current with GPM-12):**
```
ID                   TITLE                                              TYPE       STATUS         
-----------------------------------------------------------------------------------------------
GPM-12               Fix column alignment in pm list for long t...     bug        backlog
```

**After (with lipgloss table):**
```
┌──────────┬─────────────────────────────────────────────────┬──────┬─────────┐
│ ID       │ TITLE                                           │ TYPE │ STATUS  │
├──────────┼─────────────────────────────────────────────────┼──────┼─────────┤
│ GPM-12   │ Fix column alignment in pm list for long ti...  │ bug  │ backlog │
│ GPM-13   │ Make ticket IDs case-insensitive in commands    │ bug  │ backlog │
└──────────┴─────────────────────────────────────────────────┴──────┴─────────┘
```

Or borderless style:
```
ID        TITLE                                            TYPE  STATUS
GPM-12    Fix column alignment in pm list for long ti...   bug   backlog
GPM-13    Make ticket IDs case-insensitive in commands     bug   backlog
```

## Acceptance Criteria

- [ ] Tables automatically adjust to terminal width
- [ ] Unicode and emoji rendered correctly
- [ ] Color coding for status (optional, configurable)
- [ ] All list-style commands use consistent formatting
- [ ] `--no-color` flag works for script/CI usage
- [ ] Performance acceptable for 1000+ tickets

## Future Enhancements

- Export to markdown tables (`pm list --format markdown`)
- Export to CSV (`pm list --format csv`)
- Interactive table with sorting (via bubbletea TUI)
- Terminal width detection and column prioritization

