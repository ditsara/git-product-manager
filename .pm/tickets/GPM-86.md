---
assignee: ""
blocks: []
created_at: "2026-04-05T09:30:17Z"
depends_on:
    - GPM-82
    - GPM-83
id: GPM-86
labels: []
parent: GPM-81
points: 0
priority: medium
related: []
status: done
title: Deduplicate and consolidate LLM guidance in AGENTS.md
type: task
updated_at: "2026-04-05T10:02:16Z"
---

# Description

Get this repo's LLM context files into shape now that `pm ai guide` and
`.pm/AGENTS.md` exist. This is a one-time editorial task specific to the
GPM repo itself.

**Rule of thumb:**
- Content that applies to *any* repo using GPM → belongs in `pm ai guide`
- Content specific to *this* repo (GPM's own codebase) → belongs in
  `AGENTS.md`

## Steps

**1. Audit and deduplicate `AGENTS.md`**

Read through `AGENTS.md` and remove any content that is already covered
(or should be covered) by `pm ai guide` sections. Workflow instructions,
ticket schema, command references — if it belongs in the guide, remove it
from `AGENTS.md` and ensure `pm ai guide` covers it adequately (see
GPM-84).

**2. Merge `CACHE_DATA_MODEL.md` into `AGENTS.md`**

`CACHE_DATA_MODEL.md` (625 lines) documents GPM's SQLite cache schema and
sync strategy — relevant context for contributors to *this* repo. Merge the
content into `AGENTS.md` under a dedicated section, then delete
`CACHE_DATA_MODEL.md`.

**3. Run `pm ai init --for copilot`**

Once GPM-83 is implemented, run this command to add the pointer to
`.pm/AGENTS.md` in `.github/copilot-instructions.md`.

> Note: step 3 requires GPM-83 to be done first. Steps 1 and 2 can be done
> as soon as GPM-82 ships.

## Acceptance Criteria

- [ ] `AGENTS.md` contains no content duplicated in `pm ai guide`
- [ ] `AGENTS.md` contains the cache data model content (merged from
      `CACHE_DATA_MODEL.md`)
- [ ] `CACHE_DATA_MODEL.md` is deleted
- [ ] `.github/copilot-instructions.md` points to `.pm/AGENTS.md`
      (after GPM-83)
