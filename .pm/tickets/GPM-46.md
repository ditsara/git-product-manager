---
id: GPM-46
title: "Cache sync: detect and warn on broken relationship symmetry"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 3  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: GPM-2  # Parent epic or story
depends_on: [GPM-17]  # Must complete these first
blocks: []  # This blocks these tickets
related: [GPM-45]  # Related work (duplicates, see-also)

labels: [cache, validation, data-integrity]  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-08T14:38:53Z"
updated_at: "2026-02-08T14:38:53Z"
---

# Description

[Claude Sonnet 4.5]

Detect broken relationship symmetry during cache sync (when users manually edit ticket files) and warn users about inconsistencies. Optionally provide auto-healing.

## Problem Statement

While `pm link` and `pm unlink` maintain automatic symmetry, users can manually edit `.pm/tickets/*.md` files and break the inverse relationship:

**Example of broken symmetry:**
```yaml
# GPM-5.md
depends_on: [GPM-10]

# GPM-10.md  
blocks: []  # Should contain GPM-5 but doesn't!
```

This creates confusing behavior:
- `pm show GPM-5` says "depends on GPM-10"
- `pm blocked GPM-10` doesn't show GPM-5 as blocked
- Dependency graph is incomplete

## Solution: Detect During Cache Sync

When syncing the cache (triggered by lazy sync), validate relationship symmetry:

1. **Scan all tickets** for depends_on and blocks arrays
2. **Build expected inverse map**: If A depends_on B, expect B blocks A
3. **Detect mismatches**: Flag when inverse relationship is missing
4. **Report to user**: Print warnings about broken symmetry

## Implementation Approach

### Detection Logic (in `internal/cache/sync.go`)

```go
func validateRelationshipSymmetry(tickets []ticket.Ticket) []string {
    var warnings []string
    
    // Build map of expected blocks relationships
    expectedBlocks := make(map[string][]string)
    for _, t := range tickets {
        for _, dep := range t.DependsOn {
            expectedBlocks[dep] = append(expectedBlocks[dep], t.ID)
        }
    }
    
    // Check each ticket's blocks array matches expectations
    for _, t := range tickets {
        expected := expectedBlocks[t.ID]
        actual := t.Blocks
        
        // Find missing entries
        missing := difference(expected, actual)
        if len(missing) > 0 {
            warnings = append(warnings, 
                fmt.Sprintf("⚠ %s blocks array incomplete (missing: %v)", 
                    t.ID, missing))
        }
        
        // Find extra entries (orphaned blocks)
        extra := difference(actual, expected)
        if len(extra) > 0 {
            warnings = append(warnings, 
                fmt.Sprintf("⚠ %s blocks array has orphaned entries: %v", 
                    t.ID, extra))
        }
    }
    
    return warnings
}
```

### User Experience

**During normal operation (no issues):**
```bash
pm list
# No warnings, everything works normally
```

**When symmetry is broken:**
```bash
pm list
⚠ Detected relationship inconsistencies:
  • GPM-10 blocks array incomplete (missing: [GPM-5])
  • GPM-5 depends_on GPM-10, but GPM-10 doesn't block GPM-5

Run 'pm repair --fix-symmetry' to auto-heal, or edit tickets manually.

ID    TITLE                  TYPE    STATUS
...
```

## Implementation Steps

- [ ] Add `validateRelationshipSymmetry()` to `internal/cache/sync.go`
- [ ] Call validation during cache sync after parsing all tickets
- [ ] Print warnings to stderr if inconsistencies detected
- [ ] Store warnings in cache metadata table for later retrieval
- [ ] Add `pm repair --fix-symmetry` flag (optional) to auto-heal:
  - [ ] Read all tickets
  - [ ] Rebuild correct blocks arrays from depends_on
  - [ ] Write corrected YAML back to files
  - [ ] Report what was fixed

## Auto-Healing Behavior (Optional)

**Conservative approach** (recommended for first iteration):
- **Detect and warn only**: Don't automatically fix, let user decide
- **Explicit repair**: User runs `pm repair --fix-symmetry` if desired

**Aggressive approach** (future enhancement):
- **Auto-heal on sync**: Silently fix broken symmetry during cache rebuild
- **Log changes**: Print what was auto-fixed
- **Risky**: User might not expect files to be modified

## Edge Cases

- **Ticket referenced doesn't exist**: Separate validation (not symmetry issue)
- **Circular dependencies**: Separate validation (should be caught earlier)
- **Multiple issues**: Report all warnings, not just first one
- **User manually wants asymmetric data**: Too bad - system enforces correctness

## Testing

- [ ] Unit test: Detect missing blocks entries
- [ ] Unit test: Detect orphaned blocks entries
- [ ] Unit test: No warnings when symmetry is correct
- [ ] Integration test: Manually break symmetry, verify warning appears
- [ ] Integration test: `pm repair --fix-symmetry` heals broken relationships
- [ ] Test with multiple broken relationships

## Acceptance Criteria

- [ ] Broken symmetry is detected during cache sync
- [ ] Clear warnings printed to user
- [ ] Warnings include which tickets and what's wrong
- [ ] `pm repair --fix-symmetry` auto-heals (optional)
- [ ] Tests verify detection and (optional) healing
- [ ] Performance acceptable (validation runs in <100ms for 1000 tickets)# Description

