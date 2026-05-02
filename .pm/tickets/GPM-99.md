---
assignee: ""
blocks: []
created_at: "2026-05-02T10:02:03Z"
depends_on: []
id: GPM-99
labels: []
parent: ""
points: 0
priority: medium
related: []
status: todo
title: Use shell tools to show tickets if available
type: story
updated_at: "2026-05-02T10:05:26Z"
---




# Description

Use shell tools to display the output of `pm show`, if they are available.
Build the full output into a buffer first, then pipe it through the best
available pager.

## Pager call order

1. `bat` — if found in `$PATH`, invoke as `bat --language=md --paging=always`
2. `less` — if found in `$PATH`, invoke with `-R` (pass ANSI escapes through)
3. Plain `stdout` — fallback when neither tool is available

No other tools are needed; `bat` and `less` cover all realistic environments.
`more` is deliberately excluded — it is a strict subset of `less` and less
capable.

## Pager skip conditions

Skip the pager and write directly to stdout when **any** of the following are true:

- `--no-pager` flag is passed (name chosen to match `git --no-pager` convention)
- stdout is **not** a TTY (i.e. output is being piped or redirected)

## Flag

```
--no-pager    Write output directly to stdout, skipping bat/less
```

## Implementation notes

- Collect all output (YAML front matter + body + comments) into a
  `bytes.Buffer` before deciding on the display path.
- Use `os.LookPath` to detect `bat` and `less`.
- Launch the chosen pager via `exec.Command`; pipe the buffer to its stdin,
  and connect its stdout/stderr to `os.Stdout`/`os.Stderr`.
- Detect TTY with `golang.org/x/term` (`term.IsTerminal(int(os.Stdout.Fd()))`).
  The package is already a transitive dependency via `glamour`.
- The `--no-comments` flag must still be respected alongside `--no-pager`.

## Acceptance criteria

- `pm show GPM-1` opens output in `bat` when bat is installed and stdout is a TTY.
- `pm show GPM-1` falls back to `less` when bat is absent.
- `pm show GPM-1` falls back to plain stdout when neither is installed.
- `pm show GPM-1 | cat` outputs plain text (no pager) because stdout is not a TTY.
- `pm show GPM-1 --no-pager` outputs plain text regardless of TTY state.
- `pm show GPM-1 --no-comments --no-pager` still suppresses comments.
