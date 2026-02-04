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