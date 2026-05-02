---
assignee: ""
blocks: []
created_at: "2026-05-02T07:08:55Z"
depends_on: []
id: GPM-97
labels:
    - milestones
    - cli
parent: ""
points: 3
priority: medium
related: []
status: done
title: pm milestone add/remove commands with --cascade flag
type: story
updated_at: "2026-05-02T07:22:36Z"
---


# Description

Add dedicated `pm milestone add` and `pm milestone remove` subcommands for
assigning and unassigning tickets from milestones.

Currently, milestone assignment requires `pm edit TICKET-ID --field
milestones=a,b,c` which has **replace semantics** — the entire array must be
re-declared on every change. This is error-prone and makes it impossible to
add/remove a single milestone without first reading the ticket.

## Commands

### `pm milestone add <milestone-id> <ticket-id> [--cascade]`

Appends the milestone to the ticket's `milestones` field (idempotent — no-ops
if already present).

With `--cascade`: also adds the milestone to all direct and indirect child
tickets (tickets whose `parent` field resolves to this ticket recursively).

```bash
pm milestone add sprint-1 GPM-5
pm milestone add sprint-1 GPM-5 --cascade   # includes all descendants
```

### `pm milestone remove <milestone-id> <ticket-id> [--cascade]`

Removes the milestone from the ticket's `milestones` field (no-op if not
present).

With `--cascade`: also removes the milestone from all descendant tickets.

```bash
pm milestone remove sprint-1 GPM-5
pm milestone remove sprint-1 GPM-5 --cascade
```

## Behaviour Details

- **Append/remove semantics** — unlike `pm edit`, these commands never touch
  other milestones on the ticket.
- **Idempotent** — running `add` twice or `remove` on a ticket that isn't in
  the milestone is not an error.
- **Cascade is opt-in** — default behaviour is single-ticket only; `--cascade`
  must be explicit.
- **Cascade depth** — walks the full descendant tree (not just direct
  children).
- **Cascade output** — print each modified ticket ID so the user can see what
  changed.
- **Validation** — milestone ID must exist; ticket ID must exist. Errors are
  surfaced per the standard error format.
- **Git staging** — files are not staged automatically; left to the user,
  consistent with all other `pm` commands (`edit`, `move`, `assign`, etc.).
  Note: the original spec said `git add` would be called, but this was removed
  during implementation after reviewing that no other command does this.

## Acceptance Criteria

- [x] `pm milestone add sprint-1 GPM-5` adds `sprint-1` to GPM-5's milestones
  without affecting other milestones on the ticket
- [x] Running `add` a second time is a no-op (no file write, no error)
- [x] `pm milestone remove sprint-1 GPM-5` removes `sprint-1`; other milestones
  are untouched
- [x] Running `remove` when milestone is not assigned is a no-op
- [x] `--cascade` on `add` modifies all descendants and prints each ticket ID
- [x] `--cascade` on `remove` removes from all descendants and prints each
  ticket ID
- [x] Invalid milestone ID prints a clear error
- [x] Invalid ticket ID prints a clear error
- [x] Modified files are NOT auto-staged (left to user, consistent with rest of tool)
- [x] Integration tests cover single-ticket and cascade paths for both add and
  remove

