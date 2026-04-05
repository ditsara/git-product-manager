---
assignee: ""
blocks: []
created_at: "2026-04-05T08:54:17Z"
depends_on:
    - GPM-79
id: GPM-80
labels: []
parent: GPM-78
points: 0
priority: medium
related: []
status: done
title: Warn when assigning to user not in member list
type: task
updated_at: "2026-04-05T10:29:23Z"
---


# Description

After GPM-79 adds the `members` list to project config, `pm assign` should warn the user when they assign a ticket to someone who is not on that list. The warning is non-fatal — the assignment proceeds regardless.

## Motivation

The members list is opt-in. If a team has configured it, an unknown assignee is probably a typo or a new team member who hasn't been added yet. A helpful warning closes that gap without being annoying.

## Behaviour

- If `members` is **empty or absent**: no warning, no change in behaviour.
- If `members` is **non-empty** and the resolved assignee (after domain suffix is applied, per GPM-29) is **not in the list**: print a warning to stderr, then continue.
- If the assignee **is** in the list: proceed silently.

Warning message:

```
⚠ Warning: 'alice' is not in the project member list.
  To add them, edit .pm/config/project.yaml and add to the members list.
```

## Implementation

In `cmd/pm/assign.go`, after resolving the final `username` value (post domain-suffix logic from GPM-29), add:

```go
project, err := config.LoadProject(getPmPath())
if err == nil && len(project.Members) > 0 {
    found := false
    for _, m := range project.Members {
        if m == username {
            found = true
            break
        }
    }
    if !found {
        fmt.Fprintf(os.Stderr, "⚠ Warning: %q is not in the project member list.\n", username)
        fmt.Fprintf(os.Stderr, "  To add them, edit .pm/config/project.yaml and add to the members list.\n")
    }
}
```

If `LoadProject` fails (e.g., project.yaml doesn't exist), silently skip the check.

## Acceptance Criteria

- [ ] No warning when `members` list is absent or empty
- [ ] Warning printed to stderr (not stdout) when assignee not in list
- [ ] Assignment still succeeds even when warning is shown
- [ ] Warning not shown when assignee is in the list
- [ ] `make test` passes

## Testing

- Integration: assign to a listed member → no warning
- Integration: assign to an unlisted user with non-empty members list → warning on stderr
- Integration: assign with empty/absent members list → no warning

