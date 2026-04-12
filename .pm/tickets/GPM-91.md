---
assignee: ""
blocks: []
created_at: "2026-04-12T10:29:34Z"
depends_on: []
id: GPM-91
labels: []
parent: GPM-2
points: 2
priority: medium
related: []
status: done
title: Use table format for pm search
type: story
updated_at: "2026-04-12T10:44:14Z"
---


# Description

Replace the current free-form output in `pm search` with the shared
`renderTable` / `TableColumn` infrastructure already used by `pm list` and `pm
milestone list`.

## Column Layout

| # | Header | Min Width | Notes |
|---|--------|-----------|-------|
| 0 | ID     | 15        | fixed |
| 1 | TITLE  | 20        | fixed |
| 2 | MATCH  | —         | **expanding column** (absorbs spare terminal width) |
| 3 | TYPE   | 10        | fixed |
| 4 | STATUS | 15        | color-coded via existing `statusColors` palette |

ID column behavior should be consistent with `pm list`.

The MATCH column shows the body snippet (e.g. `...auth-related refactoring...`), or is blank for ID/title-only matches.

## Behaviour Changes

- Header line (`Search results for "..." (N matches):`) is retained above the table.
- The trailing blank line between results is removed (table rows are self-contained).
- `No results for "..."` plain-text message is unchanged.
- `renderTable` is called with `statusColIndex=4`, `expandCol=2`.

## Implementation Notes

- In `cmd/pm/search.go`, replace the `fmt.Printf` result loop with a `[][]string` rows slice fed to `renderTable`.
- `SearchResult.Snippet` maps to the MATCH cell; empty string renders as blank (already handled by `padRight`).
- No changes to `internal/cache/query.go` required.

## Acceptance Criteria

- [x] `pm search` output uses the same table chrome as `pm list`
- [x] MATCH column expands to fill terminal width
- [x] STATUS column is color-coded using the shared palette
- [x] Blank MATCH cell for ID/title matches (no snippet)
- [x] Header count line still printed above the table
- [x] All existing search integration tests pass
- [x] Add/update tests to assert tabular column headers (ID, TITLE, MATCH, TYPE, STATUS)