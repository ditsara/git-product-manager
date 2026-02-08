---
assignee: ""
blocks: []
created_at: "2026-02-08T07:57:09Z"
depends_on:
    - GPM-22
id: GPM-29
labels:
    - cli
    - config
    - ux
    - stage-2
parent: ""
points: 2
priority: low
related:
    - GPM-22
status: backlog
title: Add optional domain suffix configuration for assignees
type: task
updated_at: "2026-02-08T08:11:20Z"
---


# Description

[Haiku 4.5]

Add optional configuration for automatic domain suffix appending to assignees. This enhancement improves UX by reducing typing when all assignees share a common email domain.

## User Story

As a team lead at a company with a standard email domain, I want to configure a default domain suffix so that `pm assign GPM-1 alice` automatically becomes `alice@mycompany.com`, avoiding repetitive typing while remaining flexible for non-email assignees.

## Solution Approach

Add optional configuration in `.pm/config/project.yaml`:

```yaml
prefix: TICKET
assignee_domain: ""  # Optional, e.g., "mycompany.com"
```

When `assignee_domain` is configured:
1. If user provides `alice`, append domain: `alice@mycompany.com`
2. If user provides `alice@otherdomain.com`, use as-is (don't double-append)
3. If `assignee_domain` is empty, use username as-is

## Implementation Steps

- [ ] Add `assignee_domain: ""` field to project config
- [ ] Update config loading in `internal/config/project.go`
- [ ] Create helper function `appendDomain(username string, domain string) string` in `internal/ticket/fields.go`
  - Returns `username` if domain is empty
  - Returns `username` unchanged if already contains `@`
  - Returns `username@domain` otherwise
- [ ] Update `cmd/pm/assign.go` to use domain helper after validating username
- [ ] Update `pm init` to include commented example of domain configuration
- [ ] Update help text for `pm assign` to mention domain suffix behavior

## Testing Requirements

### Unit Tests
- [ ] Domain appending: no domain configured → return username as-is
- [ ] Domain appending: domain configured, username without `@` → append domain
- [ ] Domain appending: domain configured, username with `@` → return unchanged
- [ ] Domain appending: empty domain string → return username as-is
- [ ] Config loading: domain field optional and defaults to empty

### Integration Tests
- [ ] Init project with default config (no domain)
- [ ] Assign with no domain configured → username unchanged
- [ ] Edit project config to add domain
- [ ] Assign with domain configured → domain appended
- [ ] Assign full email with domain configured → unchanged

## Acceptance Criteria

- [ ] Optional `assignee_domain` field in `.pm/config/project.yaml`
- [ ] Automatic domain appending when configured
- [ ] Smart appending: skip if `@` already present
- [ ] Help text documents domain behavior
- [ ] All tests pass

## Edge Cases

- **Empty domain**: Ignored; username used as-is
- **Username with `@`**: Never modify; use as provided
- **Partial email like `alice@`**: Append domain anyway (edge case, user's responsibility)
- **Multiple `@` signs**: Should not happen; treat as valid email and skip appending
- **Config change**: Next assign will use new domain (no need to re-assign existing)

## Implementation Notes

Keep this small and focused. The domain helper is pure and testable, and integrates cleanly with existing `pm assign` logic.
