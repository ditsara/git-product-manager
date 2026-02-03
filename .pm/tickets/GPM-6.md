---
id: GPM-6
title: "Fix findTicketByID to use exact match instead of prefix"
type: bug
status: backlog
priority: medium
points: 0
parent: ""
depends_on: []
blocks: []
related: []
labels: []
assignee: ""
created_at: "2026-02-03T03:26:09Z"
updated_at: "2026-02-03T03:26:09Z"
---

# Description

[Sonnet 4.5]

**Bug:** `findTicketByID()` uses prefix matching (`strings.HasPrefix`) instead of exact ID match, which can select the wrong ticket when IDs share prefixes.

## Problem Scenario

```bash
# Repository has: PROJ-1.md, PROJ-10.md, PROJ-100.md
pm show PROJ-1
# Could match PROJ-1, PROJ-10, or PROJ-100 depending on directory order
```

## Current Behavior

```go
for _, f := range files {
    if !f.IsDir() && strings.HasPrefix(f.Name(), idPrefix) {
        return filepath.Join(ticketsPath, f.Name())
    }
}
```

Returns the **first** file that matches the prefix, which is non-deterministic.

## Solution

1. **Exact match first:** Look for `{id}.md` exactly
2. **Fallback to prefix:** Only if no exact match found (for convenience)
3. **Ambiguity detection:** If multiple prefix matches, ask user to be more specific

## Implementation

```go
func findTicketByID(id string) string {
    ticketsPath := ".pm/tickets"
    exactPath := filepath.Join(ticketsPath, id+".md")
    
    // Try exact match first
    if _, err := os.Stat(exactPath); err == nil {
        return exactPath
    }
    
    // Fallback: find by prefix, but detect ambiguity
    files, _ := os.ReadDir(ticketsPath)
    var matches []string
    for _, f := range files {
        if strings.HasPrefix(f.Name(), id) {
            matches = append(matches, f.Name())
        }
    }
    
    if len(matches) == 0 {
        fmt.Printf("Error: ticket '%s' not found\n", id)
        os.Exit(1)
    }
    if len(matches) > 1 {
        fmt.Printf("Error: ambiguous ID '%s'. Matches: %v\n", id, matches)
        os.Exit(1)
    }
    
    return filepath.Join(ticketsPath, matches[0])
}
```

## Acceptance Criteria

- [ ] `pm show PROJ-1` always shows PROJ-1, never PROJ-10
- [ ] `pm show PROJ-1` with multiple matches shows error listing options
- [ ] Prefix matching still works for convenience (e.g., `pm show PROJ-1` works if only PROJ-10 exists)
- [ ] All commands using `findTicketByID` are updated
- [ ] Tests cover exact match, prefix match, and ambiguity scenarios

