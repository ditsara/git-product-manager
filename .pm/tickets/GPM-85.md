---
id: GPM-85
title: "Add guide content: markdown formatting conventions for ticket descriptions"
type: task
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "GPM-81"  # Parent epic or story
depends_on:
  - GPM-82
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-04-05T09:17:31Z"
updated_at: "2026-04-05T09:17:31Z"
---

# Description

Add a section (or extend an existing one) in the embedded guide content to
teach LLMs how to format ticket description Markdown correctly. Without this,
LLMs tend to write long lines that are hard to read in a terminal or diff.

## What to Document

Add to `internal/guide/schema.md` (or `principles.md` if it fits better)
under a **"Writing Ticket Descriptions"** heading:

### Line wrapping

Wrap prose at **80 columns**. This keeps descriptions readable in terminals,
`git diff`, and `pm show` output without horizontal scrolling.

### Table alignment

Align table column text for readability. Long cell values that would push a
row past 80 columns are acceptable — don't sacrifice alignment just to hit
the limit.

Example (from GPM-81's sub-ticket table):

```markdown
<!-- Before: unaligned, long lines -->
| Ticket | Title | Notes |
|--------|-------|-------|
| GPM-82 | Reorganize `pm guide` → `pm ai guide` | Clean rename, no backward compat needed |
| GPM-83 | Add `pm ai init` bootstrap command | Writes stub to Claude/Copilot/Cursor/Aider targets |

<!-- After: aligned columns, long rows accepted as-is -->
| Ticket | Title                                 | Notes                                    |
|--------|---------------------------------------|------------------------------------------|
| GPM-82 | Reorganize `pm guide` → `pm ai guide` | Clean rename, no backward compat needed  |
| GPM-83 | Add `pm ai init` bootstrap command    | Writes stub to Claude/Copilot/Cursor/Aider targets |
```

### Lists and code blocks

- Indent continuation lines to align with the text, not the bullet marker
- Wrap list item prose at 80 cols; keep code blocks unwrapped

## Acceptance Criteria

- [ ] Guide includes a "Writing Ticket Descriptions" section with the above
      conventions
- [ ] 80-column wrap rule is stated explicitly
- [ ] Table alignment guidance is stated explicitly
- [ ] An example table is included to show the expected style

