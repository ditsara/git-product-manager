---
assignee: ""
blocks: []
created_at: "2026-05-02T09:32:11Z"
depends_on: []
id: GPM-98
labels:
    - cli
    - llm
    - dx
parent: ""
points: 3
priority: medium
related:
    - GPM-82
    - GPM-83
status: backlog
title: Utility to prettify and wrap markdown
type: story
updated_at: "2026-05-02T10:00:28Z"
---

# Description

[Claude Sonnet 4.6]

LLMs tend to output markdown body content as single long lines or inconsistently
wrapped paragraphs. This is hard to read in editors with no soft-wrap and creates
noisy git diffs. `pm ai tool mdfmt` provides a fast, YAML-aware formatter that
LLMs can call immediately after editing a ticket to normalize its body.

## Usage

```
pm ai tool mdfmt GPM-1          # print formatted content to STDOUT
pm ai tool mdfmt GPM-1 --write  # rewrite the ticket file in-place
```

## Library Research

No existing Go library was found that satisfies all three requirements: markdown
input, configurable column-width paragraph wrapping, and markdown output. The
full landscape was researched:

| Library | Stars | Wraps at N cols? | Notes |
|---------|-------|-----------------|-------|
| `shurcooL/markdownfmt` | 817 | ❌ | De-facto "gofmt for markdown"; collapses paragraphs to one long line. No YAML front matter support. Open wrap requests since 2015, never implemented. |
| `Kunde21/markdownfmt` | 61 | ❌ | Active Goldmark-based fork used by `bwplotka/mdox`. `WithSoftWraps()` preserves existing breaks, does not add new ones. |
| `teekennedy/goldmark-markdown` | 19 | ❌ | Goldmark renderer that outputs markdown instead of HTML. Style options (heading, indent, thematic break) but no column width. |
| `mdigger/goldmark-formatter` | 9 | ❌ | Similar Goldmark markdown renderer. No wrap. |
| `muesli/reflow` | 770 | ✅ | Already a transitive dep (via charmbracelet). Word-wraps plain text at N cols, but **completely markdown-unaware** — would wrap inside code blocks and break table rows. |
| `Gosayram/go-mdfmt` | 1 | ✅ | Only Go lib with paragraph reflow at configurable width. Created June 2025; uses byte-length instead of rune-width, custom regex parser with no established markdown lib. Not suitable. |

**Goldmark was also evaluated** as a parse-then-re-render pipeline
(`teekennedy/goldmark-markdown`). It can output markdown but has no word-wrap
option — the wrap would still need to be written from scratch, while also
accepting normalization side effects (list marker style, heading style). Not
worth the two extra dependencies.

**Conclusion:** Implement `internal/mdfmt` from scratch. The custom state
machine is the right approach — there is no existing wheel.

### Algorithm Reference: flowmark (Python, MIT)

Python's [`flowmark`](https://github.com/jlevy/flowmark) (actively maintained,
designed specifically for LLM/git workflows) uses a two-layer approach that is
directly portable to Go and should guide this implementation:

**Layer 1 — Block-level state machine:** Walk the source line by line.
Identify block-level constructs (fenced code, headings, thematic breaks, block
quotes, lists, tables, HTML blocks) and pass them through verbatim. Only
plain prose paragraphs are sent to the word-wrapper.

**Layer 2 — Inline-aware word wrap (the key insight from flowmark):**
Before word-wrapping a paragraph's text, extract "atomic constructs" — inline
elements that must never be split across lines — replacing them with numbered
placeholders. After wrapping, restore the originals. This ensures a long
markdown link or inline code span is never broken mid-token.

Atomic constructs to protect:
- Inline code spans: `` `code` ``, ` ``multi-backtick`` `
- Markdown links: `[text](url)`, `[text][ref]`
- Autolinks: `<https://example.com>`

Additionally, if word-wrapping causes a word to land at the start of a new
line and that word looks like a markdown block marker (`-`, `*`, `+`, `>`,
`#`, or `1.`), it must be backslash-escaped to prevent re-interpretation.

**mdformat** (Python, MIT) was also reviewed. Its `--wrap INTEGER` feature
works correctly but is tightly coupled to markdown-it's AST token model (~23KB
renderer). Not suitable for porting; the flowmark approach is the right one.

### Solution Approach

### CLI Structure

Create a two-level subcommand group under `pm ai`:

```
pm ai tool            # parent group: "Utility tools for AI/LLM use"
pm ai tool mdfmt      # the formatter
```

New files:
- `cmd/pm/ai_tool.go` — defines `aiToolCmd` and adds it to `aiCmd`
- `cmd/pm/ai_tool_mdfmt.go` — defines `aiToolMdfmtCmd`
- `internal/mdfmt/mdfmt.go` — formatter logic
- `internal/mdfmt/mdfmt_test.go` — unit tests

### Formatter Behaviour

The formatter is implemented as a pure Go function in `internal/mdfmt`:

```go
// Format rewraps the markdown body of a ticket file at wrapWidth columns.
// YAML front matter (delimited by leading and trailing ---) is passed through
// unchanged. Returns the complete formatted file content.
func Format(src []byte, wrapWidth int) ([]byte, error)
```

**YAML front matter:** Detected by the same `---`/`---` delimiter logic used
in `internal/ticket`. The front matter block is copied verbatim; only the body
after the closing `---` is reformatted.

**Block-level pass-through (no reflowing):**
- Fenced code blocks (` ``` ` or `~~~` delimited) — content and fences
  preserved exactly, including long lines
- Indented code blocks (4-space / tab indent)
- ATX headings (`#`, `##`, …)
- Setext heading underlines (`===`, `---`)
- Blank lines (structure preserved)
- Thematic breaks (`---`, `***`, `___`)
- Block quotes (`> ` lines) — treat as opaque in v1
- List items — treat as opaque in v1
- HTML blocks — pass through verbatim
- Tables — pass through verbatim

**Paragraph rewrapping:** Consecutive non-blank lines that do not trigger any
block-level rule above are collected into a single paragraph buffer and
reflowed to `wrapWidth` columns (default: 80) using the two-step inline-aware
algorithm:

1. **Extract atomic constructs** — scan for inline code spans and markdown
   links using regex; replace each with a placeholder token (`\x00N\x00`).
2. **Word-wrap** — split on whitespace, walk tokens tracking column position,
   emit a newline when the next word would exceed `wrapWidth`. Words are never
   split mid-token.
3. **Escape markdown specials at line starts** — if wrapping causes a token
   that looks like a block marker (`-`, `*`, `+`, `>`, `#…`, `N.`, `N)`) to
   land at the start of a new line, prepend a backslash.
4. **Restore placeholders** — substitute placeholders back for original inline
   constructs.

The formatter is intentionally conservative: when in doubt about whether a
line is "prose", leave it alone. False negatives (failing to rewrap) are
acceptable; false positives (mangling structure) are not.

### Ticket File Resolution

`mdfmt` takes a ticket ID (e.g. `GPM-1`). Resolution follows the same
case-insensitive logic used by `pm show`: look up the ticket path from the
cache (or fall back to scanning `.pm/tickets/`). Return a clear error if the
ID is not found.

### Output Modes

- Default (no flags): write formatted content to STDOUT. Exit 0.
- `--write`: overwrite the ticket file in-place. Print
  `✓ GPM-1 formatted` to STDOUT. Exit 0 on success.

### Guide Update

Add a short entry to `internal/guide/workflow.md` in the **"After Editing
Tickets"** note (or a new "Formatting" section near the end) instructing the
LLM to run `pm ai tool mdfmt TICKET-ID --write` after editing ticket body
content.

## Edge Cases

| Case | Expected behaviour |
|------|--------------------|
| Ticket has no YAML front matter | Treat entire content as markdown body; format normally |
| Ticket body is already wrapped correctly | Output is identical to input (idempotent) |
| Ticket contains only front matter, empty body | Output unchanged |
| Fenced code block with long lines | Lines inside block are never truncated |
| Inline code span spanning a word boundary | Entire span stays on one line; never broken |
| Markdown link `[text](url)` near column 80 | Entire link token stays on one line; never broken |
| Wrapped line starts with `-`, `*`, `+`, `>`, `#`, or `N.` | Token is backslash-escaped to prevent block re-interpretation |
| `--write` on a read-only file | Return a clear error: `Error: cannot write GPM-1: <os error>` |
| Ticket ID not found | `Error: ticket GPM-99 not found` |
| Paragraph word longer than wrapWidth | Emit the word on its own line rather than exceeding the limit |

## Implementation Steps

- [ ] Create `cmd/pm/ai_tool.go` with `aiToolCmd` parent command
- [ ] Create `cmd/pm/ai_tool_mdfmt.go` with `aiToolMdfmtCmd`, `--write` flag,
      ticket-path resolution, and STDOUT/write-back output logic
- [ ] Implement `internal/mdfmt/mdfmt.go`:
  - `Format(src []byte, wrapWidth int) ([]byte, error)` — top-level entry point
  - `extractAtomics(s string) (placeholders map[string]string, out string)` —
    replace inline code spans and markdown links with `\x00N\x00` placeholders
  - `restoreAtomics(s string, placeholders map[string]string) string`
  - `wrapParagraph(text string, width int) string` — word-wrap using
    placeholder-protected tokens; escape markdown specials at line starts
  - Block-level state machine that calls `wrapParagraph` only for prose lines
- [ ] Write `internal/mdfmt/mdfmt_test.go` covering:
  - Plain paragraph wrap and idempotency
  - Inline code span not broken across lines
  - Markdown link not broken across lines
  - Block marker at wrapped line start is escaped
  - Fenced code block passthrough (including long lines)
  - YAML front matter passthrough
  - Empty body
  - Word longer than wrapWidth
- [ ] Update `internal/guide/workflow.md` with formatting instruction
- [ ] Build (`make build`) and smoke-test manually
- [ ] Run `go test ./...` — all tests pass

## Acceptance Criteria

- [ ] `pm ai tool mdfmt GPM-X` prints formatted content to STDOUT; YAML front
      matter is byte-for-byte identical to the original
- [ ] `pm ai tool mdfmt GPM-X --write` rewrites the file and prints
      `✓ GPM-X formatted`
- [ ] Running `mdfmt` twice on the same file produces identical output
      (idempotent)
- [ ] Fenced code blocks are never modified
- [ ] Inline code spans and markdown links are never split across wrapped lines
- [ ] `pm ai guide workflow` mentions `pm ai tool mdfmt`
- [ ] All existing tests continue to pass
