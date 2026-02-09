---
id: GPM-58
title: "Fix migration files not found after go install"
type: bug
status: backlog  # Current workflow state
priority: high  # low, medium, high, critical
points: 5  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""  # Parent epic or story
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: [build, installation, migration]  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-09T15:48:29Z"
updated_at: "2026-02-09T15:48:29Z"
---

# Description

[Claude Sonnet 4.5]

Users who install the `pm` binary via `go install` encounter a critical error when running `pm init`:

```
Error: could not find migration files
```

This prevents new users from initializing the tool after following the documented installation instructions. The issue occurs because `go install` compiles only the binary without including the `migrations/` directory that contains the SQL schema files.

## Problem Analysis

**Current Implementation:**
- Migration files exist as external SQL files in `migrations/` directory
- `cache.FindMigrationPath()` searches for migrations on the filesystem
- Searches relative to: current dir, executable location, parent dirs
- Fails when binary is installed to `$GOPATH/bin` with no migrations nearby

**Root Cause:**
- `go install` builds a standalone binary in `$GOPATH/bin`
- No mechanism to copy migration files alongside the binary
- The lookup logic assumes migrations are in a nearby directory

**Impact:**
- New users cannot initialize the tool after `go install`
- Breaks the documented Quick Start workflow
- Only workaround is building from source with `./bin/pm`

## Solution Approach

**Embed migration files into the binary using Go's `embed` package.**

This makes the binary completely standalone with no external file dependencies. The migration SQL files will be compiled into the binary itself at build time.

### Why This Approach?

✅ **Zero configuration** - Works with `go install` out of the box  
✅ **Single binary** - No separate data files to manage or distribute  
✅ **Version consistency** - Migrations are locked to the code version  
✅ **Simpler deployment** - Just copy/install the binary  
✅ **Development friendly** - Embed reads actual files at build time  

### Technical Implementation

Use `github.com/golang-migrate/migrate/v4/source/iofs` driver to read migrations from an embedded filesystem instead of disk files.

**Key changes:**
- Add `//go:embed migrations/*.sql` directive to embed files
- Switch from `file://` source to `iofs` source for golang-migrate
- Remove `FindMigrationPath()` function (no longer needed)
- Simplify `RunMigrations()` to always use embedded migrations

## Implementation Steps

- [ ] **Add embed directive** in `internal/cache/migrate.go`
  - Import `embed` package
  - Add `//go:embed` directive to include `../../migrations/*.sql` files
  - Create `fs.FS` variable with embedded migration files
  - Note: Path is relative to the Go source file location

- [ ] **Update migration initialization**
  - Replace `file://` source with `iofs.New()` using embedded filesystem
  - Import `github.com/golang-migrate/migrate/v4/source/iofs`
  - Remove `migrationPath` parameter from `RunMigrations()`

- [ ] **Update `EnsureCacheReady()` function** in `ensure.go`
  - Remove call to `FindMigrationPath()`
  - Remove migration path validation logic
  - Pass only `dbPath` to `RunMigrations()`
  - Simplify error handling (no "migrations not found" case)

- [ ] **Update `cmd/pm/init.go`**
  - Remove migration path lookup logic (lines 73-78)
  - Remove `migrationPath` variable
  - Call `cache.RunMigrations(dbPath)` with only database path
  - Simplify error messages

- [ ] **Clean up unused code**
  - Delete `FindMigrationPath()` function from `ensure.go`
  - Remove all migration path discovery logic
  - Clean up imports if any become unused

- [ ] **Update all tests**
  - Update `migrate_test.go` - remove migration path setup
  - Update `sync_test.go` - remove migration path variables  
  - Update `ensure_test.go` (if exists) - remove migration path handling
  - All tests should use embedded migrations automatically

- [ ] **Test the fix locally**
  - Build with `./scripts/build.sh` - verify it works
  - Run `./bin/pm init sandbox --prefix TEST`
  - Verify database is created successfully
  - Run integration tests: `go test ./...`

- [ ] **Test via go install**
  - Uninstall any existing version: `rm $(which pm)` (if exists)
  - Install from local source: `go install ./cmd/pm`
  - Run `pm init test-dir --prefix TEST` in a new directory
  - Verify it succeeds without "migration files not found" error

- [ ] **Update documentation** (if needed)
  - Check AGENTS.md for references to migration file locations
  - Verify README.md installation instructions still accurate
  - Add release notes about embedded migrations

## Acceptance Criteria

- [ ] User runs: `go install github.com/ditsara/git-product-manager/cmd/pm@latest`
- [ ] User runs: `pm init . --prefix TEST` in a new directory
- [ ] Command succeeds without "could not find migration files" error
- [ ] Database is created at `.pm/.cache.db` with correct schema
- [ ] All existing integration tests pass
- [ ] No external migration files required for installed binary
- [ ] Development workflow still works (build from source)

## Technical Notes

### Go embed directive syntax

```go
//go:embed ../../migrations/*.sql
var migrationsFS embed.FS
```

The path is relative to the `.go` file containing the directive. Since `migrate.go` is in `internal/cache/`, we need `../../migrations/` to reach the project root.

### Migration source change

```go
// Before:
m, err := migrate.New("file://" + migrationPath, "sqlite3://" + dbPath)

// After:
migrationSource, err := iofs.New(migrationsFS, "migrations")
if err != nil {
    return fmt.Errorf("failed to create migration source: %w", err)
}
m, err := migrate.NewWithSourceInstance("iofs", migrationSource, "sqlite3://" + dbPath)
```

### Required imports

```go
import (
    "embed"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/source/iofs"
    _ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)
```

## Edge Cases

1. **Development vs Production:** Embed reads actual files at compile time, so changes to migration files require recompilation (expected)

2. **Backward Compatibility:** Existing local builds continue to work - embedded migrations are used transparently

3. **Migration Versioning:** golang-migrate handles `schema_migrations` table identically regardless of source type

4. **File Paths:** Embedded paths use forward slashes even on Windows (Go normalizes automatically)

## Risks

- **Low risk:** `embed` is a standard library feature since Go 1.16
- **No breaking changes:** CLI interface stays the same, only internal implementation changes
- **Well tested:** Existing tests validate behavior, we're just changing how migrations are loaded

