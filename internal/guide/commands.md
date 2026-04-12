# GPM Commands Reference

## Setup

```bash
pm init . --prefix PROJ          # Initialise .pm/ in current directory
pm ai guide                      # Display this full guide
pm ai guide <section>            # Display one section (workflow|schema|commands|principles)
pm ai init                       # Append a GPM pointer to LLM tool config files
pm ai init --for claude          # Append only to CLAUDE.md
```

> **Important:** Never create new tickets or edit ticket yaml directly. Use `pm
> new`, `pm edit`, and `pm move` to manage tickets — the CLI handles IDs,
> validation, and git tracking. Editing ticket markdown content directly is
> okay.

## Tickets

```bash
pm new "Title"                   # Create a new story ticket
pm new --type task "Title"       # Create a task (types: epic|story|task|bug)
pm new --type bug "Title" -P GPM-5  # Create a bug as child of GPM-5
pm show GPM-42                   # Show full ticket (rendered Markdown + metadata)
pm show GPM-42 --no-comments     # Show ticket without comments
pm list                          # List all open tickets
pm list --all                    # Include done/canceled tickets
pm list --status in-progress     # Filter by status
pm list --assignee alice         # Filter by assignee
pm list --label auth             # Filter by label
pm list --parent GPM-5           # Show children of GPM-5
pm list --parent GPM-5 --all     # Include closed children
pm list --milestone sprint-1     # Show tickets in a milestone
pm edit GPM-42                   # Open ticket in $EDITOR
pm edit GPM-42 --field assignee=alice      # Update a single field
pm edit GPM-42 --field labels=bug,api      # Update array field (replaces)
pm edit GPM-42 --field milestones=sprint-1 # Assign to milestone
pm move GPM-42 in-progress       # Change ticket status
pm move GPM-42 done              # Mark ticket done
```

## Comments & History

```bash
pm comment GPM-42 -m "LGTM"     # Add a comment
pm comment GPM-42                # Open $EDITOR to write a comment
pm history GPM-42                # Show status change history (from git)
```

## Relationships

```bash
pm link GPM-42 GPM-10 --type depends-on   # GPM-42 depends on GPM-10
pm link GPM-42 GPM-10 --type parent       # GPM-42 is a child of GPM-10
pm link GPM-42 GPM-10 --type blocks       # GPM-42 blocks GPM-10
pm link GPM-42 GPM-10 --type related      # Loose association
pm unlink GPM-42 GPM-10                   # Remove relationship
pm blocked                                 # Show all tickets with unresolved deps
pm blocked GPM-42                          # Show what blocks/is blocked by GPM-42
```

## Milestones

```bash
pm milestone create "Sprint 1" --due 2026-06-30   # Create milestone
pm milestone create "Sprint 1" --id sprint-1       # Create with explicit ID
pm milestone list                                   # List all milestones
pm milestone list --state closed                    # Show only closed milestones
pm milestone list --overdue                         # Show overdue milestones
pm milestone list --with-progress                   # Show completion %
pm milestone show sprint-1                          # Show milestone details + progress
pm milestone close sprint-1                         # Close milestone
pm milestone close sprint-1 --force                 # Close even if tickets unfinished
```

## Assign

```bash
pm assign GPM-42 alice           # Shorthand for pm edit --field assignee=alice
```

## AI Integration

```bash
pm ai guide                      # Read development workflow guidance
pm ai guide workflow             # Workflow section
pm ai guide schema               # Ticket schema reference
pm ai guide commands             # This commands reference
pm ai guide principles           # Key principles
pm ai init                       # Bootstrap all supported LLM tool config files
pm ai init --for claude          # Bootstrap only CLAUDE.md
```
