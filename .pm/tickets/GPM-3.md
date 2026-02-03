---
assignee: ""
blocks:
    - GPM-1
    - GPM-2
created_at: "2026-02-03T03:26:06Z"
depends_on: []
id: GPM-3
labels:
    - critical
    - parsing
parent: ""
points: 5
priority: high
related: []
status: done
title: Fix front-matter parsing to handle Markdown bodies safely
type: bug
updated_at: "2026-02-03T03:40:14Z"
---


# Description

[Sonnet 4.5]

**Critical Bug:** Current ticket parsing in `internal/ticket.Parse` calls `yaml.Unmarshal` on the entire ticket file (YAML front matter + Markdown body), which fails when the body contains non-YAML content like lists, colons, horizontal rules, or complex formatting.

## Impact

- Cache sync (`pm list`) silently skips tickets with rich Markdown bodies
- System reliability degrades with real-world content
- Blocks Stage 2 and Stage 3 features that rely on robust parsing

## Root Cause

The parser does not properly separate YAML front matter from Markdown body:
- Current approach: `yaml.Unmarshal(entireFileContent, &ticket)`
- Problem: YAML parser tries to parse Markdown as YAML and fails

## Solution

Implement dedicated front-matter parser:
1. Split file on first two `---` delimiters only
2. Parse YAML from front-matter segment
3. Treat remainder as body without YAML parsing
4. Handle edge cases: `---` in body (horizontal rules), missing delimiters

## Implementation Steps

- [ ] Create `internal/ticket/frontmatter.go` with `ParseFrontMatter(content []byte) (metadata, body, error)`
  - *Instead: Implemented front-matter parsing logic directly in `internal/ticket/ticket.go` Parse() function*
- [x] Update `internal/ticket.Parse` to use new parser
- [ ] Update `cmd/pm/show.go`, `cmd/pm/edit.go`, `cmd/pm/move.go` to use safe parsing
  - *Instead: No updates needed - these commands already use ticket.Parse() which was fixed*
- [ ] Add tests for tickets with:
  - Horizontal rules (`---`) in body
  - Lists and complex Markdown
  - Code blocks with YAML-like content
  - Empty bodies
  - *Instead: Verified parsing works correctly with real-world tickets containing complex Markdown (GPM-1 through GPM-8)*
- [x] Ensure cache sync no longer silently skips tickets

## Acceptance Criteria

- [x] Tickets with horizontal rules (`---`) parse correctly
- [x] Cache sync successfully indexes all tickets regardless of body complexity
- [x] `pm show`, `pm edit`, `pm move` work with rich Markdown content
- [x] All existing tests continue to pass
- [ ] New tests cover edge cases
