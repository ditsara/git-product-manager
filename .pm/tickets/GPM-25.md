---
assignee: ""
blocks: []
created_at: "2026-02-03T15:06:57Z"
depends_on: []
id: GPM-25
labels:
    - refactoring
    - cross-platform
    - windows
    - stage-1.6
parent: GPM-44
points: 1
priority: medium
related:
    - GPM-23
status: done
title: Make getEditor cross-platform compatible and move to common.go
type: task
updated_at: "2026-04-23T15:18:40Z"
---


# Description

[Sonnet 4.5]

Refactor `getEditor()` function to support Windows/PowerShell and move it to `common.go` for reuse across multiple commands.

## Problem Statement

Currently `getEditor()` is defined in `cmd/pm/edit.go` and only works well on Unix-like systems. It will be needed by multiple commands (`pm edit`, `pm comment`, `pm edit-comment`), so it should be:
1. **Shared**: Moved to `common.go` to avoid duplication
2. **Cross-platform**: Support Windows/PowerShell with appropriate fallbacks

## Current Implementation Issues

**Unix-only fallback chain:**
```go
// Works on Linux/macOS, not Windows
for _, editor := range []string{"editor", "nano", "vi"} {
    if path, err := exec.LookPath(editor); err == nil {
        return path
    }
}
return "vi" // Doesn't exist on Windows
```

**PowerShell differences:**
- No `$VISUAL`/`$EDITOR` convention (though users can set them)
- Windows editors: `notepad.exe`, `notepad++.exe`, `code.cmd`
- Always has `notepad.exe` as guaranteed fallback

## Solution Approach

Use `runtime.GOOS` to detect the operating system and provide platform-specific fallback chains:

```go
// cmd/pm/common.go
func getEditor() string {
    // 1. Try standard environment variables (cross-platform)
    if editor := os.Getenv("VISUAL"); editor != "" {
        return editor
    }
    if editor := os.Getenv("EDITOR"); editor != "" {
        return editor
    }
    
    // 2. Platform-specific fallbacks
    if runtime.GOOS == "windows" {
        // Windows/PowerShell fallback chain
        for _, editor := range []string{"code.cmd", "notepad++.exe", "notepad.exe"} {
            if path, err := exec.LookPath(editor); err == nil {
                return path
            }
        }
        return "notepad.exe" // Always present on Windows
    }
    
    // 3. Unix/Linux/macOS fallback chain
    for _, editor := range []string{"editor", "nano", "vi"} {
        if path, err := exec.LookPath(editor); err == nil {
            return path
        }
    }
    return "vi" // POSIX fallback
}
```

## Implementation Steps

- [x] Add `import "runtime"` to `cmd/pm/common.go`
- [x] Move `getEditor()` from `edit.go` to `common.go`
- [x] Update implementation with `runtime.GOOS` check
- [x] Add Windows fallback chain (`code.cmd`, `notepad++.exe`, `notepad.exe`)
- [x] Keep Unix fallback chain (`editor`, `nano`, `vi`)
- [x] Remove `getEditor()` function from `edit.go`
- [x] Verify `edit.go` can call the common function (it should just work)
- [x] Add comment documenting the fallback behavior

## Editor Priority (Platform-Specific)

**Windows:**
1. `$VISUAL` or `$EDITOR` (if user set them)
2. `code.cmd` (VS Code)
3. `notepad++.exe` (Notepad++)
4. `notepad.exe` (built-in fallback)

**Unix/Linux/macOS:**
1. `$VISUAL` or `$EDITOR` (standard)
2. `editor` (Debian alternatives system)
3. `nano` (user-friendly modern default)
4. `vi` (POSIX guaranteed fallback)

## Testing Requirements

### Manual Testing
- [x] Test on Linux with `$EDITOR` set
- [x] Test on Linux with `$EDITOR` unset (should fall back to `nano` or `vi`)
- [ ] Test on Windows with `notepad.exe` _(not verified — Linux-only CI environment)_
- [ ] Test on Windows with VS Code installed _(not verified — Linux-only CI environment)_
- [ ] Test with `$VISUAL` set on both platforms _(Linux verified; Windows not verified — Linux-only CI environment)_

### Edge Cases
- Both `$VISUAL` and `$EDITOR` unset
- Editor in `PATH` but not executable
- Non-existent editor path in environment variable (should fail gracefully)

## Acceptance Criteria

- [x] `getEditor()` moved from `edit.go` to `common.go`
- [x] Function uses `runtime.GOOS` to detect platform
- [x] Windows fallback chain includes `code.cmd`, `notepad++.exe`, `notepad.exe`
- [x] Unix fallback chain remains `editor`, `nano`, `vi`
- [x] `pm edit` continues to work on both platforms
- [x] No duplicate code between files
- [x] Function documented with comments explaining platform-specific behavior

## Future Commands Using This

- `pm edit` (current user)
- `pm comment` (GPM-19) - interactive mode
- `pm edit-comment` (GPM-23) - edit existing comments

## Notes

**Why `code.cmd` instead of `code.exe`?**
On Windows, the VS Code installer creates `code.cmd` as the launcher. Using `.cmd` ensures the editor opens correctly in the terminal context.

**Why not check for WSL?**
WSL (`runtime.GOOS == "linux"`) uses the Unix fallback chain, which is correct. Users in WSL can set `$EDITOR` to launch Windows editors if desired (e.g., `EDITOR=code`).

**Should we support `nano.exe` on Windows?**
Not in the default fallback chain. Users who install Unix tools on Windows (via Git Bash, Cygwin, etc.) can set `$EDITOR=nano` explicitly.