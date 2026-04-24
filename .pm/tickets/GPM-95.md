---
assignee: ""
blocks: []
created_at: "2026-04-24T00:25:15Z"
depends_on: []
id: GPM-95
labels: []
parent: ""
points: 0
priority: medium
related: []
status: backlog
title: Do not show broken symmetry warnings for closed tickets (see GPM-46)
type: story
updated_at: "2026-04-24T00:29:59Z"
---


# Description

## Problem

`validateRelationshipSymmetry` (in `internal/cache/validate.go`) currently warns
on every broken pair regardless of ticket status. This generates noise for
`done`/`canceled` tickets — a closed ticket that was linked before being closed
often has a stale relationship that is harmless and uninteresting.

Example: GPM-10 is `done`, GPM-5 depends_on GPM-10 but GPM-10 doesn't block
GPM-5. Warning fires every `pm list`, but nobody cares — GPM-10 is closed.

## Desired Behaviour

- **Definition of closed**: ticket status in the `completed` group in `workflow.yaml`
- **Both tickets closed**: suppress the warning entirely.
- **One ticket open, one ticket closed**: show the warning only for the open side.
  i.e., only emit the warning when `fromTicket` (the ticket with the declaration)
  is open.
- **Both tickets open**: warning fires as before.

## Implementation Notes

`relationshipData` currently carries only `fromTicket`, `toTicket`, and
`relType`. It does **not** carry status. The fix requires one of:

1. Add a `status` field to `relationshipData` and populate it in
   `scanTicketFiles` alongside the existing relationship collection loop, OR
2. Pass the `[]ticketData` slice into `validateRelationshipSymmetry` so it can
   build a `statusByID` map for the lookup.

Option 2 is simpler (no struct change) and preferred. The signature becomes:

```go
func validateRelationshipSymmetry(tickets []ticketData, relationships []relationshipData) []string
```

Inside the function, build `statusByID map[string]string` from `tickets`, then
before appending any warning check that the `fromTicket` status is not `done` or
`canceled`.

Update the call site in `sync.go` accordingly.

## Acceptance Criteria

- [ ] Warnings are suppressed when `fromTicket` is `done` or `canceled`
- [ ] Warnings are suppressed when both tickets are closed
- [ ] Warning still fires when `fromTicket` is open, regardless of `toTicket` status
- [ ] `validateRelationshipSymmetry` signature updated to accept `[]ticketData`
- [ ] Unit tests in `internal/cache/validate_test.go` cover:
  - both tickets closed → no warning
  - from-ticket closed, to-ticket open → no warning
  - from-ticket open, to-ticket closed → warning emitted
  - both open → warning emitted (existing behaviour preserved)
- [ ] Integration test: break symmetry on a `done` ticket, verify no warning on `pm list`
