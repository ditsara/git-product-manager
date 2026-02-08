---
id: GPM-40
title: "Error messages missing trailing newline"
type: bug
status: backlog
priority: low
points: 1

parent: ""
depends_on: []
blocks: []
related: []

labels: [ux, cli, polish]
assignee: ""
created_at: "2026-02-08T09:26:21Z"
updated_at: "2026-02-08T09:26:21Z"
---

# Description

[Claude Sonnet 4.5]

When commands fail due to validation errors (missing arguments, invalid flags, etc.), the error output does not include a trailing newline. This causes the shell prompt to appear on the same line as the error message, creating visual clutter.

## Example

```bash
$ pm comment 
Error: accepts 1 arg(s), received 0
Usage:
  pm comment <ticket-id> [flags]

Flags:
      --amend              Edit an existing comment instead of creating a new one
  -a, --author string      Author name (defaults to git user.name)
  -h, --help               help for comment
  -m, --message string     Comment text (skips editor if provided)
      --timestamp string   Specific comment timestamp to edit (RFC3339 format)

Whoops. There was an error while executing your CLI 'accepts 1 arg(s), received 0'(master ?:1 ✗) sandbox 
```

Notice the shell prompt `(master ?:1 ✗) sandbox` appears immediately after the last line of output with no newline separator.

## Expected Behavior

Error output should end with a newline, so the shell prompt appears on its own line:

```bash
$ pm comment 
Error: accepts 1 arg(s), received 0
Usage:
  pm comment <ticket-id> [flags]

Flags:
      --amend              Edit an existing comment instead of creating a new one
  -a, --author string      Author name (defaults to git user.name)
  -h, --help               help for comment
  -m, --message string     Comment text (skips editor if provided)
      --timestamp string   Specific comment timestamp to edit (RFC3339 format)

(master ?:1 ✗) sandbox 
```

## Root Cause

This is likely happening because:
1. Cobra's default error handling prints errors without a trailing newline
2. Our custom error formatting (`fmt.Fprintf(os.Stderr, "Error: %v", err)`) doesn't include `\n`
3. We're not using Cobra's `SilenceUsage` correctly

## Implementation Tasks

- [ ] Audit all error output paths in `cmd/pm/*.go`
- [ ] Ensure all error messages written to stderr include a trailing newline
- [ ] Check Cobra command configuration for proper error handling
- [ ] Test with various error conditions:
  - [ ] Missing required arguments (`pm comment`)
  - [ ] Invalid flags (`pm list --invalid`)
  - [ ] File not found errors (`pm show NONEXISTENT`)
  - [ ] Validation failures (`pm move TICKET invalid-status`)
  - [ ] Git errors (operation outside git repo)

## Solution Approach

### Option 1: Fix Individual Error Sites
Add `\n` to each error message:
```go
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
```

### Option 2: Centralized Error Handler
Use Cobra's `PersistentPreRunE` or custom error wrapper:
```go
func printError(err error) {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}
```

### Option 3: Use Cobra's Built-in Error Handling
Configure Cobra properly and let it handle formatting:
```go
rootCmd.SilenceUsage = false
rootCmd.SilenceErrors = false
```

**Recommendation:** Start with Option 2 (centralized handler) for consistency, then audit individual sites.

## Testing

Manual testing required - automated tests don't easily capture terminal formatting:

```bash
# Test each command's error cases
./bin/pm comment               # Missing arg
./bin/pm move                  # Missing args
./bin/pm show FAKE-123         # Invalid ticket
./bin/pm list --invalid        # Invalid flag
./bin/pm new                   # Missing title
```

For each, verify:
- Error message is clear
- Shell prompt appears on new line
- No extra blank lines

## Acceptance Criteria

- [ ] All error messages end with a newline
- [ ] Shell prompt always appears on its own line after errors
- [ ] No regression in error message clarity or formatting
- [ ] All common error paths tested manually

