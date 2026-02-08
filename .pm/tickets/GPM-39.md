---
id: GPM-39
title: "Implement state_groups in workflow.yaml and filter completed tickets by default"
type: story
status: backlog
priority: high
points: 5

parent: ""
depends_on: []
blocks: []
related: []

labels: [filtering, workflow, ux, enhancement]
assignee: ""
created_at: "2026-02-08T09:21:26Z"
updated_at: "2026-02-08T09:21:26Z"
---

# Description

[Claude Sonnet 4.5]

Implement the `state_groups` feature from the AGENTS.md specification to enable semantic grouping of workflow states. Use this to make `pm list` exclude completed tickets by default, reducing clutter while maintaining flexibility for custom workflows.

## Problem Statement

Currently, `pm list` shows all top-level tickets regardless of status, including those marked as "done". As projects grow, the list becomes cluttered with completed work. Users want to focus on active work by default.

However, workflow states are user-defined in `workflow.yaml`, so we can't hard-code assumptions about which states represent "completed" work. Different teams use different workflows (kanban, scrum, custom).

## User Stories

**As a project manager**, I want `pm list` to hide completed tickets by default, so I can focus on active work without manual filtering.

**As a developer with a custom workflow**, I want to define which states are considered "completed" in my workflow.yaml, so the tool adapts to my process.

**As a new user**, I want clear documentation in the default workflow.yaml explaining states, initial_state, and state_groups, so I understand how to customize my workflow.

## Solution: State Groups

Implement the `state_groups` feature already designed in AGENTS.md:

```yaml
# Workflow states define the lifecycle of tickets
# States are labels - no enforced transitions (teams move tickets freely)
states:
  - backlog      # Not yet planned
  - todo         # Ready to start
  - in-progress  # Currently being worked on
  - done         # Completed work

# initial_state: The default status for new tickets created via 'pm new'
initial_state: backlog

# state_groups: Semantic groupings for filtering and reporting
# These groups are optional but enable smart default behavior
state_groups:
  # Active work - shown by default in pm list
  active: [todo, in-progress]
  
  # Completed work - hidden by default in pm list
  completed: [done]
  
  # Everything not yet done
  incomplete: [backlog, todo, in-progress]
```

## Implementation Steps

### 1. Update workflow.yaml Template
- [ ] Add comprehensive comments to `cmd/pm/templates/workflow.yaml`
- [ ] Explain each section: states, initial_state, state_groups
- [ ] Include the default state_groups (active, completed, incomplete)
- [ ] Provide examples of when to customize

### 2. Parse state_groups in Config
- [ ] Update `internal/config/workflow.go` to parse state_groups
- [ ] Add `StateGroups map[string][]string` to Workflow struct
- [ ] Validate that states in groups exist in the states list
- [ ] Handle missing state_groups gracefully (fall back to showing all)

### 3. Add Filtering Flags to pm list
- [ ] Add `--completed` flag: Show only tickets in "completed" states
- [ ] Add `--incomplete` flag: Show only tickets NOT in "completed" states
- [ ] Add `--active` flag: Show only tickets in "active" states
- [ ] Update help text with examples

### 4. Change Default Behavior (Breaking Change)
- [ ] Make `pm list` exclude "completed" state_group by default
- [ ] Document this in help text: "By default, hides completed tickets. Use --all to show everything."
- [ ] If workflow.yaml has no "completed" group, show all tickets (backward compatible)
- [ ] Update integration tests for new default behavior

### 5. Update Documentation
- [ ] Update AGENTS.md if needed (spec already has this)
- [ ] Add migration note for existing users
- [ ] Update pm list help text with new filtering examples

## Technical Design

### Config Parsing (internal/config/workflow.go)

```go
type Workflow struct {
    States       []string              `yaml:"states"`
    InitialState string                `yaml:"initial_state"`
    StateGroups  map[string][]string   `yaml:"state_groups"`
}

func (w *Workflow) Validate() error {
    // Existing validation...
    
    // Validate state_groups reference existing states
    for groupName, states := range w.StateGroups {
        for _, state := range states {
            if !contains(w.States, state) {
                return fmt.Errorf("state_group '%s' references unknown state '%s'", groupName, state)
            }
        }
    }
    return nil
}

func (w *Workflow) IsCompleted(status string) bool {
    if completedStates, ok := w.StateGroups["completed"]; ok {
        return contains(completedStates, status)
    }
    return false // No completed group = nothing is completed
}
```

### List Command (cmd/pm/list.go)

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List tickets",
    Long: `Lists tickets with optional filtering.

By default, shows top-level incomplete tickets (hides completed work).

Examples:
  pm list                      # Top-level incomplete tickets
  pm list --all                # All tickets (including completed)
  pm list --completed          # Only completed tickets
  pm list --active             # Only active work (todo, in-progress)
  pm list --status done        # Specific status filter`,
    Run: func(cmd *cobra.Command, args []string) {
        // Load workflow to get state_groups
        workflow, err := config.LoadWorkflow(".pm")
        
        // Determine which states to exclude
        var excludeStates []string
        if !showAll && !showCompleted && !showActive {
            // Default: exclude completed states
            if completedStates, ok := workflow.StateGroups["completed"]; ok {
                excludeStates = completedStates
            }
        }
        
        // Build query with exclusion...
    },
}

func init() {
    listCmd.Flags().Bool("all", false, "Show all tickets (including completed)")
    listCmd.Flags().Bool("completed", false, "Show only completed tickets")
    listCmd.Flags().Bool("incomplete", false, "Show only incomplete tickets")
    listCmd.Flags().Bool("active", false, "Show only active work")
    listCmd.Flags().String("status", "", "Filter by specific status")
    // ... existing flags
}
```

### SQL Query Logic

```go
// Add WHERE clause to exclude completed states
if len(excludeStates) > 0 {
    placeholders := strings.Repeat("?,", len(excludeStates))
    placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
    
    if strings.Contains(query, "WHERE") {
        query += " AND status NOT IN (" + placeholders + ")"
    } else {
        query += " WHERE status NOT IN (" + placeholders + ")"
    }
    
    for _, state := range excludeStates {
        queryArgs = append(queryArgs, state)
    }
}
```

## Edge Cases & Backward Compatibility

### No state_groups Defined
If workflow.yaml doesn't have state_groups (old format or minimal setup):
- Default behavior: Show all tickets (backward compatible)
- Flags like `--completed` do nothing (log warning)

### Empty completed Group
```yaml
state_groups:
  completed: []
```
Result: Nothing is hidden by default (explicit choice)

### Missing Groups
If user defines `state_groups` but omits "completed":
- Default behavior: Show all tickets (no completed group = nothing to hide)

### Flag Conflicts
- `--all` overrides `--completed` / `--active` / `--incomplete`
- `--status done` overrides group filtering (explicit takes precedence)
- Multiple group flags: Last one wins

## Testing Requirements

### Unit Tests (workflow_test.go)
- [ ] Parse state_groups from YAML
- [ ] Validate states in groups exist
- [ ] IsCompleted() returns correct results
- [ ] Handle missing state_groups gracefully

### Integration Tests (integration_list_test.go)
- [ ] `pm list` hides completed tickets by default
- [ ] `pm list --all` shows everything
- [ ] `pm list --completed` shows only done work
- [ ] `pm list --active` shows only active states
- [ ] Works with --parent filtering
- [ ] Backward compatible with workflow.yaml without state_groups

### Migration Test
- [ ] Existing projects without state_groups continue to work
- [ ] Users can opt into new behavior by adding state_groups

## Acceptance Criteria

- [ ] Default workflow.yaml has comprehensive comments explaining states, initial_state, and state_groups
- [ ] state_groups parsed and validated in config loading
- [ ] `pm list` excludes "completed" tickets by default (if group is defined)
- [ ] `--all`, `--completed`, `--active`, `--incomplete` flags work correctly
- [ ] Backward compatible: projects without state_groups show all tickets
- [ ] All tests pass
- [ ] Help text updated with examples
- [ ] No breaking changes for users who don't use state_groups

## Breaking Changes & Migration

**Breaking Change**: `pm list` default behavior changes for projects that add state_groups to workflow.yaml.

**Migration Path**:
1. Existing projects without state_groups: No change (show all tickets)
2. New projects: Get state_groups by default, completed tickets hidden
3. Users who want old behavior: Use `pm list --all` or remove state_groups

**Communication**: Add to CHANGELOG and update documentation.

## Future Enhancements

- `pm list --blocked` (if "blocked" state_group exists)
- Burndown charts using state_groups
- Velocity tracking (tickets moved to "completed" per week)
- Custom group queries: `pm list --group my-custom-group`

