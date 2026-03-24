# GPM Development Workflow

This project uses **Git Product Manager (GPM)** — a ticket-driven development workflow
designed for LLM-human collaboration. Follow this process for every change.

## The Workflow

### 1. Idea / Request
The user provides a feature request, bug report, or improvement idea. It may be informal
("we should fix X") or detailed. It may reference existing code or be completely new.

### 2. Ticket Creation
Create a comprehensive ticket using `pm new`:

```bash
pm new --type task "Implement feature X"
```

A good ticket includes:
- **LLM Attribution**: Add `[Model Name]` at the start of the description.
  **CRITICAL:** Use your actual model identifier — do NOT assume. Check your system
  prompt. Examples: `[Claude Sonnet 4.6]`, `[Claude Haiku 4.5]`, `[GPT-4.1]`.
- **Problem Statement**: Clear description of the issue or feature
- **Solution Approach**: Recommended implementation strategy
- **Edge Cases**: Explicitly defined behaviour for corner cases
- **Implementation Steps**: Actionable checklist (`- [ ] ...`)
- **Acceptance Criteria**: Clear, testable outcomes (`- [ ] ...`)
- **Code Examples**: When helpful, include example code or expected behaviour

Tickets should be detailed enough that another LLM could implement them without
additional context.

### 3. Review & Refinement
Present the ticket to the user before writing any code. Iterate until both parties
agree on the specification. Clarify ambiguities — never assume implementation details.

If the user asks "what happens if X?", document the answer in the ticket before
implementing. Edge cases often become test cases.

### 4. Implementation
Follow the implementation steps checklist:
- Check off items (`- [x]`) as they are completed
- Write tests for edge cases identified in the ticket
- Build and run tests after each meaningful change (`go test ./...`)
- Update the ticket if implementation reveals new considerations

### 5. Verification & Completion
- Verify all acceptance criteria are met
- Check off all acceptance criteria (`- [x]`)
- Mark the ticket done: `pm move TICKET-ID done`
- Add notes about what was done differently than planned (if applicable)
- **Do not commit to git without the user's explicit approval**

## Key Principles

- **Tickets are the specification** — don't start coding before the ticket is well-defined
- **Clarify before coding** — ask rather than assume; rework is expensive
- **Incremental verification** — build and test after every change, not in batches
- **User owns git history** — never commit or push without the user's explicit instruction

## Planning Documentation

- **For ticket-specific work:** Document planning notes directly in the ticket's
  markdown file (add an `## Implementation Planning` section after the description).
  Keep the plan co-located with the specification for future reference.
- **For cross-ticket or project-wide analysis:** Use a separate session file.
