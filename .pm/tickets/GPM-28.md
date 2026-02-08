---
id: GPM-28
title: "Embed Development Workflow Guidance into CLI"
type: epic
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""  # Parent epic (for nested epics)
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
assignee: ""  # GitHub username or email
created_at: "2026-02-04T05:30:49Z"
updated_at: "2026-02-04T05:30:49Z"
---

# Description

When other LLMs work on repos using GPM, they won't have access to AGENTS.md. This epic covers embedding the essential development workflow guidance directly into the CLI and project config, making GPM self-documenting for any AI assistant.

## Problem

- LLMs working on GPM-based repos lack context about the development workflow
- AGENTS.md is in the git-product-manager repo, not in repos where GPM is *used*
- Without workflow guidance, LLMs may make suboptimal decisions about ticket structure and implementation approach

## Solution Approach

Embed workflow guidance in two complementary ways:
1. **CLI command** (`pm guide`): Display workflow principles and best practices
2. **Config file** (`.pm/config/WORKFLOW_GUIDE.md`): Versioned, customizable guidance created during `pm init`

This makes guidance discoverable and self-contained in each repo.

## Related Tickets

- **Potential sub-tasks:**
  - `pm guide` command implementation
  - WORKFLOW_GUIDE.md template creation
  - Help system enhancements
  - Documentation updates

## Demonstrated Workflow Pattern (From Actual Implementation)

Based on successful session implementing GPM-45, here's the proven workflow pattern for future LLM assistants:

### Phase 1: Specification & Clarification
1. **Read the Epic/Ticket**: Thoroughly understand the problem statement and acceptance criteria
2. **Identify Clarifications Needed**: Don't assume implementation details—ask for explicit decisions
3. **Design First**: Document the approach (data structures, algorithm, symmetry logic, error handling) before coding
4. **User Agreement**: Get user confirmation on key design decisions before implementation begins

**Example from GPM-45**: Tickets specified "automatic symmetry" but left implementation strategy open. We clarified: Should de-duplication happen on read or write? How to handle idempotency? What about atomic rollback on failure? User reviewed options, we chose Normalize-on-write with atomic rollback. This prevented wasted coding.

### Phase 2: Implementation with Continuous Verification
1. **Incremental Coding**: Create one file at a time, build immediately after
2. **Run Tests After Each Change**: Don't batch changes; verify immediately
3. **Check Off Checklist Items**: Mark implementation steps complete as you finish them
4. **Update Timestamps**: When modifying tickets, refresh `updated_at` to reflect current state

**Example from GPM-45**: 
   - Created relationships.go, tested it
   - Created link.go/unlink.go, tested both commands
   - Created relationships_test.go with comprehensive coverage
   - Each step verified before moving to next

### Phase 3: Completion & Documentation
1. **Verify All Acceptance Criteria**: Go through each one, confirm it's met
2. **Run Full Test Suite**: Ensure no regressions
3. **Build Final Binary**: Verify clean compilation
4. **Update Ticket Status**: Mark as "done" using `pm move`, check off all checklist items
5. **Skip Git Commit**: Let user review and commit—they own the git history

**Example from GPM-45**: All 42 tests passed, both commands working, acceptance criteria verified, ticket marked done.

### Key Principles Demonstrated

- **Tickets ARE the specification**: Don't start coding before the ticket is crystal clear
- **User review prevents rework**: Clarify ambiguities before implementation (saves hours)
- **Testing is not optional**: Run tests after every change to catch issues early
- **Checklists drive implementation**: The acceptance criteria and implementation steps guide what to build
- **Transparency matters**: Explain design choices to user for approval before coding

---

## LLM Conversation output

Previous LLM conversation output below, for reference.

I'd recommend a two-part approach:

1. New pm guide command (or pm workflow)

A dedicated command to display development workflow principles:

* Shows the essential sections from "Development Workflow" (1.5) in a consumable format
* Outputs Markdown or plain text suitable for LLM context windows
* Optional --json flag for programmatic/structured access
* Could have sub-modes like:
    * pm guide workflow - The iterative ticket-driven process
    * pm guide principles - Key principles (tickets are specs, explicit edge cases, etc.)
    * pm guide example - A sample workflow exchange
    * pm guide schema - Ticket YAML structure

2. Embedded workflow guide file

Store a .pm/config/WORKFLOW_GUIDE.md (created during pm init):

* Contains the essential development workflow sections extracted from AGENTS.md
* Lives in the repo alongside tickets, so other LLMs can access it
* Can be versioned and customized per project
* Keeps guidance close to the tickets it describes

3. Enhanced help system

Add an "For AI Assistants" section to root help:

* Brief mention that pm guide provides workflow context
* Suggests: "If you're an LLM working on this repo, run pm guide workflow"
* Reference that tickets contains actual specifications
* Note that pm show <id> displays full ticket context

Why this works:

* Offline-capable: Other repos get the guide without AGENTS.md
* Self-contained: Everything needed is in .pm
* Discoverable: Help naturally points to it
* Scalable: Can add more guides (architecture decisions, testing patterns, etc.)
* LLM-friendly: Structured, focused output without noise