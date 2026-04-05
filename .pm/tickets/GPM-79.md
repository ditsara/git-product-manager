---
id: GPM-79
title: "Add configurable member list with assignee autocomplete"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-78"  # Parent epic or story
depends_on: []  # Must complete these first
blocks:
  - GPM-80
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-04-05T08:54:17Z"
updated_at: "2026-04-05T08:54:17Z"
---

# Description

Add a configurable list of project members to `project.yaml`. This list is used as the completion source when the user presses `<TAB>` on the second argument of `pm assign`.

## Motivation

Currently `pm assign GPM-1 <TAB>` returns nothing. With a member list configured, the shell will suggest known team members, reducing typos and making discovery instant.

## Implementation

**1. Extend `internal/config/project.go`**

Add `Members []string` to the `Project` struct:

```go
type Project struct {
    Prefix         string   `yaml:"prefix"`
    AssigneeDomain string   `yaml:"assignee_domain,omitempty"`
    Members        []string `yaml:"members,omitempty"`
}
```

**2. Update `pm init` template**

Add a commented example in `createProjectConfig` (in `cmd/pm/init.go`):

```yaml
prefix: MYPROJECT
# assignee_domain: "mycompany.com"  # optional — appended when assigning without @
# members:                           # optional — used for tab completion on pm assign
#   - alice
#   - bob
```

**3. Add completion helper in `cmd/pm/completion_helpers.go`**

```go
func completeMembers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    pmPath := getPmPath()
    project, err := config.LoadProject(pmPath)
    if err != nil || len(project.Members) == 0 {
        return nil, cobra.ShellCompDirectiveNoFileComp
    }
    return project.Members, cobra.ShellCompDirectiveNoFileComp
}
```

**4. Wire into `pm assign` in `cmd/pm/assign.go`**

Update `ValidArgsFunction` to use `completeMembers` for the second argument:

```go
ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    if len(args) == 0 {
        return completeTicketIDs(cmd, args, toComplete)
    }
    if len(args) == 1 {
        return completeMembers(cmd, args, toComplete)
    }
    return nil, cobra.ShellCompDirectiveNoFileComp
},
```

## Acceptance Criteria

- [ ] `members` field in `Project` struct, optional, defaults to empty slice
- [ ] `pm init` generated `project.yaml` includes commented-out `members` example
- [ ] `pm assign GPM-1 <TAB>` completes from the `members` list when configured
- [ ] When `members` is empty or absent, completion falls back to no suggestions (no error)
- [ ] `make test` passes

## Testing

- Unit: `LoadProject` with `members` field present and absent
- Integration: configure members in project.yaml, verify `pm assign` completion returns them
