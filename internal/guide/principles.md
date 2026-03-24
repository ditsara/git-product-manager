# GPM Key Principles

Follow these principles when working on any GPM-managed repository.

## Tickets Are the Specification

- Read the ticket thoroughly before writing any code
- The implementation steps checklist defines *what* to build
- The acceptance criteria define *how to verify* it's correct
- If a ticket is ambiguous, update it and get user agreement before coding
- Tickets should be detailed enough that another LLM could implement them

## Never Commit to Git Without User Approval

**This is non-negotiable.**

Do not run `git commit`, `git push`, or any equivalent operation unless the user
explicitly asks you to. The user owns the git history. Your job is to write correct
code and update ticket files; committing and pushing is the user's decision.

When work is complete, tell the user what changed and ask if they want to commit.

## Clarify Before Coding

- Identify ambiguities in the ticket before starting implementation
- Ask the user to resolve design questions — don't assume
- Document the agreed approach in the ticket before writing code
- A few minutes of clarification prevents hours of rework

## Incremental Verification

- Build after every meaningful change: `go build ./...`
- Run tests after every change: `go test ./...`
- Do not batch multiple unverified changes together
- Fix failures immediately rather than accumulating them

## Check Off Checklists as You Work

- Mark implementation steps done as you complete them: `- [x]`
- Mark acceptance criteria done when verified: `- [x]`
- If you deviate from the plan, update the ticket to explain why
- When all criteria are met, mark the ticket done: `pm move TICKET-ID done`

## LLM Attribution

When creating or substantially updating tickets, add your model name at the start
of the description:

```
[Claude Sonnet 4.6]

Problem statement here...
```

This helps future maintainers understand the provenance of the specification.

## Transparency

- Explain design choices to the user before implementing them
- Surface trade-offs rather than silently picking an option
- If you discover a bug or missing case while implementing, create a ticket for it
  rather than silently fixing unrelated things

## Replacement Over Append

When updating array fields (`labels`, `depends_on`, `milestones`, etc.), the new value
**replaces** the entire array — it does not append. This is consistent with how single-value
fields work. Document this choice in tickets when it affects user-visible behaviour.
