---
id: GPM-13
title: "Make ticket IDs case-insensitive in commands"
type: bug
status: backlog
priority: medium
points: 2

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""
depends_on: []
blocks: []
related: [GPM-6]

labels: [ux, usability]
assignee: ""
created_at: "2026-02-03T04:07:29Z"
updated_at: "2026-02-03T04:07:29Z"
---

# Description

[Sonnet 4.5]

**Bug:** Commands like `pm show` and `pm edit` require exact case matching for ticket IDs, which is unintuitive and frustrating for users.

## Current Behavior

```bash
$ bin/pm show gpm-12
Error: ticket with ID starting with 'gpm-12' not found.

$ bin/pm show GPM-12
# Works! Shows the ticket
```

Users must remember the exact case of the ticket prefix (GPM vs gpm), which is unnecessary friction.

## Expected Behavior

Both should work:
```bash
$ pm show gpm-12    # Should work
$ pm show GPM-12    # Should work
$ pm show GpM-12    # Should work (though unlikely to be typed)
```

## Root Cause

In `cmd/pm/common.go`, `findTicketByID()` uses case-sensitive string matching:

```go
func findTicketByID(pmPath, id string) (string, error) {
    ticketsPath := filepath.Join(pmPath, "tickets")
    files, err := os.ReadDir(ticketsPath)
    // ...
    for _, file := range files {
        if strings.HasPrefix(file.Name(), id) {  // Case-sensitive!
            return filepath.Join(ticketsPath, file.Name()), nil
        }
    }
}
```

## Solution

Normalize both the search term and filenames to uppercase (or lowercase) for comparison:

```go
func findTicketByID(pmPath, id string) (string, error) {
    ticketsPath := filepath.Join(pmPath, "tickets")
    files, err := os.ReadDir(ticketsPath)
    if err != nil {
        return "", fmt.Errorf("failed to read tickets directory: %w", err)
    }

    // Normalize search ID to uppercase for comparison
    normalizedID := strings.ToUpper(id)

    for _, file := range files {
        if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
            continue
        }
        
        // Compare uppercase versions
        filename := strings.ToUpper(file.Name())
        if strings.HasPrefix(filename, normalizedID) {
            // Return the ACTUAL filename (preserving case)
            return filepath.Join(ticketsPath, file.Name()), nil
        }
    }

    return "", fmt.Errorf("ticket with ID starting with '%s' not found", id)
}
```

## Affected Commands

All commands that use `findTicketByID()`:
- `pm show <id>`
- `pm edit <id>`
- `pm move <id> <status>`
- `pm assign <id> <user>` (when implemented)
- `pm comment <id>` (when implemented)

## Implementation Steps

- [ ] Update `findTicketByID()` in `cmd/pm/common.go` to use case-insensitive matching
- [ ] Verify all commands using `findTicketByID()` now work with lowercase IDs
- [ ] Add test case: create GPM-99, verify `pm show gpm-99` works
- [ ] Update error messages to show what the user typed (preserve case in error output)

## Edge Cases to Consider

- **Multiple matches with same prefix**: Already handled by existing prefix logic
- **Filesystem case sensitivity**: 
  - Linux/Unix: Filenames are case-sensitive, so GPM-1.md ≠ gpm-1.md
  - macOS: Usually case-insensitive filesystem (GPM-1.md == gpm-1.md)
  - Windows: Case-insensitive (GPM-1.md == gpm-1.md)
  - Solution works across all platforms since we're comparing strings, not filesystem operations

## Acceptance Criteria

- [ ] `pm show gpm-1` works the same as `pm show GPM-1`
- [ ] `pm edit gpm-1` works the same as `pm edit GPM-1`
- [ ] `pm move gpm-1 done` works the same as `pm move GPM-1 done`
- [ ] Error messages still show the ID as the user typed it
- [ ] Case-insensitive matching doesn't break existing functionality
- [ ] Tests verify behavior across different case variations

## Related

Related to GPM-6 (Fix findTicketByID to use exact match) - that ticket addresses prefix matching, this addresses case sensitivity. Both improve the robustness of ID resolution.

