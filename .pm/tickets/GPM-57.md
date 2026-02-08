---
id: GPM-57
title: "Consider archival strategy for completed tickets"
type: task
status: backlog
priority: low
points: 0

parent: ""
depends_on: []
blocks: []
related: []

labels: [ux, infrastructure, future]
assignee: ""
created_at: "2026-02-08T15:34:02Z"
updated_at: "2026-02-08T15:34:02Z"
---

# Description

**Task:** Evaluate and potentially implement an archival strategy for old, completed tickets to keep the `.pm/tickets/` directory organized and the editor file tree manageable.

## Problem Statement

As the ticket count grows (currently heading toward 50+), the `.pm/tickets/` directory becomes increasingly unwieldy in the editor's file tree. Deciding whether to implement archival requires balancing:

- **UX:** Cleaner mental model of "active work" vs. "historical record"
- **Discoverability:** Less noise when browsing active tickets
- **Simplicity:** Avoiding unnecessary operational complexity
- **Reversibility:** Ensuring archived tickets can be easily unarchived if needed

## Discussion: Should We Archive?

### Arguments For Archival

- **Psychology:** Active work vs. historical record feels cleaner
- **Discoverability:** `pm list` becomes more "signal" (active tickets), less "noise"
- **Git readability:** `git log .pm/tickets/` is cleaner without historical closed tickets
- **Psychological barrier:** Archival signals "this is really done, we can think about it differently"

### Arguments Against Archival

- **Simplicity:** Everything flat is the simplest mental model
- **Git history intact anyway:** Moving files is just a commit; history is preserved
- **Filtering already works:** `pm list --status done` already hides completed tickets from daily view
- **Operational burden:** Requires archival criteria + potentially new commands + periodic maintenance
- **Low pain threshold:** Even at 100+ tickets, a flat directory is manageable for most systems
- **Querying by reference:** If a ticket references an archived one, it's still easy to find

## Potential Archival Criteria

If we implement archival, we'd need to decide on criteria:

### Option 1: Time-based (Simplest)
- Archive after 30+ days of being closed
- Pros: Automatic, deterministic, easy to explain
- Cons: Arbitrary; some tickets might benefit from staying active longer
- Implementation: `pm archive --auto-close 30`

### Option 2: Activity-based
- Archive if no changes in N months
- Pros: Respects "truly forgotten" tickets
- Cons: Slightly more complex to track

### Option 3: Manual with Suggestions
- Auto-suggest archival, user decides
- Pros: Full control, no risk of hiding something important
- Cons: Requires discipline

### Option 4: Hybrid (Recommended if Proceeding)
- Archive automatically after 60 days of no activity
- Keep easy unarchival: `pm unarchive <id>`
- Structure: `.pm/tickets/{ID}.md` → `.pm/tickets/archive/{ID}.md`
- Git history remains searchable via `git log`

## Implementation Notes (If Pursued)

- Add `archived: boolean` field to ticket YAML (minimal overhead)
- Update filters to support `--archived` / `--active` flags
- Create optional `pm archive <id>` and `pm unarchive <id>` commands
- Ensure archival is **reversible** and doesn't break references
- Cache should handle archive queries transparently

## Recommendation

**Defer this.** Current scale doesn't warrant archival complexity. Revisit when:
- `.pm/tickets/` directory has 100+ files
- User reports UX issues with editor file tree
- Filtering (`pm list --status done`) proves insufficient

At that point, implement **Option 4 (Hybrid)** with reversible archival.

## Acceptance Criteria (When Implemented)

- [ ] Decision made: implement or remain flat
- [ ] If implementing: criteria defined in writing
- [ ] If implementing: backward compatibility maintained
- [ ] If implementing: archival is reversible
- [ ] Documentation updated

