---
assignee: ""
blocks: []
created_at: "2026-02-03T03:52:16Z"
depends_on:
    - GPM-10
id: GPM-11
labels:
    - cache
    - ux
    - troubleshooting
parent: GPM-44
points: 3
priority: medium
related:
    - GPM-9
status: canceled
title: Implement pm repair command
type: task
updated_at: "2026-04-23T16:00:15Z"
---


# Description

[Sonnet 4.5]

Provide an explicit command for users to manually rebuild the cache when they encounter issues or want to troubleshoot.

## Use Cases

1. **User suspects cache is stale**: `pm repair` forces full resync
2. **Debugging**: User wants to see what's happening during cache rebuild
3. **After manual ticket edits**: Ensure cache reflects filesystem changes
4. **Troubleshooting**: User wants control instead of auto-recovery

## Command Design

```bash
# Basic usage - rebuild cache from filesystem
pm repair

# Verbose output showing each ticket processed
pm repair --verbose

# Dry-run: show what would be done
pm repair --dry-run

# Aliases for discoverability
pm rebuild-cache
pm fix
```

## Implementation

```go
var repairCmd = &cobra.Command{
    Use:   "repair",
    Aliases: []string{"rebuild-cache", "fix"},
    Short: "Rebuild the cache database from filesystem",
    Long: `Rebuilds the .cache.db by:
  1. Backing up current cache (if it exists)
  2. Deleting corrupted cache
  3. Running migrations to create fresh schema
  4. Syncing all tickets from .pm/tickets/
  
Use this when:
  - Cache seems out of sync with filesystem
  - You've manually edited ticket files
  - Debugging cache-related issues`,
    Run: func(cmd *cobra.Command, args []string) {
        verbose, _ := cmd.Flags().GetBool("verbose")
        dryRun, _ := cmd.Flags().GetBool("dry-run")
        
        pmPath := ".pm"
        dbPath := filepath.Join(pmPath, ".cache.db")
        
        // Backup existing cache
        if !dryRun {
            if _, err := os.Stat(dbPath); err == nil {
                backupPath := dbPath + ".backup"
                os.Rename(dbPath, backupPath)
                fmt.Printf("✓ Backed up cache to %s\n", backupPath)
            }
        }
        
        // Rebuild
        if dryRun {
            fmt.Println("Would delete .cache.db")
            fmt.Println("Would run migrations")
            fmt.Println("Would sync all tickets")
        } else {
            os.Remove(dbPath)
            cache.RunMigrations(dbPath, migrationPath)
            cache.SyncCache(pmPath)
            fmt.Println("✓ Cache rebuilt successfully")
        }
    },
}
```

## Output Example

```
$ pm repair --verbose
✓ Backed up cache to .pm/.cache.db.backup
✓ Deleted corrupted cache
✓ Running migrations... (2 applied)
✓ Syncing tickets from filesystem...
  • GPM-1: Stage 2: Collaboration and History
  • GPM-2: Stage 3: Advanced Relationships and Search
  • GPM-3: Fix front-matter parsing...
  ... (8 tickets synced)
✓ Cache rebuilt successfully
```

## Acceptance Criteria

- [ ] `pm repair` successfully rebuilds cache from filesystem
- [ ] Backs up existing cache before deleting (safety)
- [ ] `--verbose` flag shows detailed progress
- [ ] `--dry-run` flag shows what would happen without doing it
- [ ] Works even if `.cache.db` is completely missing
- [ ] Help text explains when and why to use this command

## Related

Works alongside GPM-9 (auto-recovery) - this is for manual control.
Depends on GPM-10 (migration check) for robust initialization.
