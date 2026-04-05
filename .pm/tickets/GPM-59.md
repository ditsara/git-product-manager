---
assignee: ""
created_at: "2026-02-09T16:27:09Z"
id: GPM-59
labels:
    - bug
    - list-command
    - database
points: 3
priority: high
status: canceled
title: 'Fix: pm list --completed --all causes SQL syntax error'
type: bug
updated_at: "2026-04-05T08:58:02Z"
---


# Description

[Claude Sonnet 4.5]

When running `pm list --completed --all`, the command fails with:
```
Error querying tickets: near "AND": syntax error
```

This prevents users from combining these two flags, even though they have a logical meaning: "show only completed tickets, and show all levels of the hierarchy."

## Problem Analysis

**Reproduction:**
```bash
cd sandbox
../bin/pm list --completed --all
# Error: near "AND": syntax error
```

**Root Cause:**
The flag logic in `cmd/pm/list.go` (lines 88-106) has a precedence issue:
1. When `--completed` is set, it populates `includeStates` with completed states
2. When `--all` is also set, the expectation should be "show completed tickets at all levels"
3. However, the current logic treats these flags independently without considering their interaction
4. This causes SQL query generation to build an invalid query

**Code Location:** `cmd/pm/list.go:88-106` (flag parsing logic) and lines 159-188 (query building)

## Expected Behavior

The command should:
- Show only completed tickets (those in `done` and `canceled` states)
- Show them at all levels of the hierarchy (not just top-level)
- Combine the filters properly into valid SQL

## Solution Approach

The fix should clarify flag precedence:
1. `--status X` - most specific, overrides all others
2. `--all` - most general, shows everything (overrides `--completed`, `--active`, `--incomplete`)
3. `--completed` - show only completed states (overrides default filtering)
4. `--active` - show only active states
5. `--incomplete` - show only incomplete states
6. Default (no flags) - show incomplete tickets only (exclude completed)

When conflicting flags are used (e.g., `--all --completed`), the more general flag should win. In this case:
- `--all --completed` → Show all tickets (--all is more general)
- `--all --active` → Show all tickets (--all is more general)
- `--completed --active` → This is contradictory, should error or use first flag

OR alternatively, interpret them as: `--all` removes level filtering, and state flags still apply:
- `--completed --all` → Show completed tickets at all levels (preferred interpretation)

## Implementation Steps

- [ ] Analyze flag precedence requirements (clarify desired behavior)
- [ ] Fix the flag logic to handle conflicting/combined flags
- [ ] Add tests for flag combinations:
  - `pm list` (default - incomplete only)
  - `pm list --all` (all tickets)
  - `pm list --completed` (completed only, top-level)
  - `pm list --completed --all` (completed only, all levels)
  - `pm list --active` (active only)
  - `pm list --active --all` (active only, all levels)
  - `pm list --incomplete` (incomplete only)
  - `pm list --status done` (specific state only)
- [ ] Ensure SQL queries are properly formed for all combinations
- [ ] Verify no "near AND" syntax errors

## Test Case

```bash
# Should work without error:
pm list --completed --all

# Should show only completed tickets at all hierarchy levels:
pm list --completed --all | grep -E "done|canceled"
```

## Edge Cases

1. **Conflicting state filters** (e.g., `--completed --active`)
   - Decide if this is an error or if one takes precedence
   
2. **State filter with parent filter**
   - `pm list --completed --parent SANDBOX-1`
   - Should show completed children under SANDBOX-1

3. **All level indicators together**
   - `pm list --all --completed --active` (nonsensical)
   - Should error or use first/last flag

## Technical Notes

- Query building happens in lines 116-190 of list.go
- State filtering is applied at lines 159-188
- The issue appears to be that includeStates gets set but showAll also affects query structure
- Need to ensure WHERE clause is not duplicated or malformed
## Resolution

**Bug no longer reproducible.** All flag combinations (`--completed --all`, `--active --all`, `--incomplete --all`, etc.) work correctly and return valid results with no SQL errors.

**Root cause fixed as a side effect of GPM-69** (commit `426bf99`): the raw SQL query builder in `list.go` was replaced wholesale with a Bob ORM implementation in `internal/cache/query.go`. The new implementation builds each query path correctly, eliminating the malformed `WHERE … AND` construction that caused this error.

Verified 2026-04-05 against current codebase — no action required.
