Here is a comprehensive technical specification for your Git-based project
management system. This document is designed to be "LLM-ready," providing the
necessary structure, logic, and context for an AI to generate a functional
prototype.

---

# Specification: Git-Native Project Management System (GPM)

## 1. Vision & Core Philosophy

The goal is to eliminate the context switch between coding and project management. All "tickets" are stored as structured data (YAML + Markdown) within the same repository as the code.

* **Single Source of Truth:** The repository contains both the "what" (tasks) and the "how" (code).
* **GitOps Workflow:** CRUD operations are performed via a CLI, and changes are committed/pushed like code.
* **Process as Code:** Workflows, labels, and validation rules are defined in version-controlled config files.
* **Auditability:** Git history serves as the immutable audit trail for ticket lifecycles.

---

## 2. File Architecture

### 2.1 Directory Structure

To keep the root clean, all data resides in a hidden `.pm/` directory.

```text
.project/
├── .pm/
│   ├── tickets/              # Flat directory of all tasks/stories
│   │   ├── XXX-101.md
│   │   └── XXX-102.md
│   ├── config/
│   │   ├── workflow.yaml     # State machine definitions
│   │   └── labels.yaml       # Allowed tags
│   └── .cache.db             # Git-ignored SQLite index for fast CLI queries
└── src/                      # Your actual application code

```

### 2.2 Ticket Schema (Hybrid Format)

Tickets use **YAML Front-Matter** for machine-readable metadata and **Markdown** for human-readable content.

**Example: `.pm/tickets/DEV-123.md**`

```yaml
---
id: DEV-123
title: "Implement OAuth2 Login"
type: story          # Options: epic, story, task, bug
status: todo         # Must match workflow.yaml
priority: high
points: 5
epic: ARCH-45        # Relationship link
labels: [auth, api]
assignee: dan_i
created_at: 2026-01-31T09:00:00Z
updated_at: 2026-01-31T09:30:00Z
---

# Description
As a user, I want to log in via Google.

## Acceptance Criteria
- [ ] Connect to Google Provider
- [ ] Store JWT in secure cookie

# Discussion
**@dan_i** (2026-01-31): Should we use a library or native fetch?

```

---

## 3. Technical Requirements

### 3.1 Uniqueness & Identification

* **K-Sorted IDs:** To prevent merge conflicts in a distributed environment, use a prefix + short hash or a K-sorted UID (e.g., `PROJ-a1b2`). Avoid sequential integers (1, 2, 3) which collide on different branches.

### 3.2 State Machine (`workflow.yaml`)

Define valid transitions to prevent "illegal" moves.

```yaml
states: [backlog, todo, in-progress, review, done]
transitions:
  todo: [in-progress]
  in-progress: [review, todo]
  review: [done, in-progress]

```

### 3.3 The "Bad YAML" Guardrail

The implementation must include a **Validation Layer**:

* **Auto-Fix:** On save, the CLI should fix minor formatting (indentation, trailing whitespace).
* **Linter:** Before commit, check for required fields, valid status, and correct date formats.
* **Git Hook:** A pre-commit hook should block the commit if `.pm/tickets/` contains invalid YAML.

---

## 4. Feature Set for Prototype

### Phase 1: The Core CLI

* `pm init`: Setup the `.pm` directory and default configs.
* `pm new "Title"`: Generate a new ticket file from a template.
* `pm list`: Scan `.pm/tickets/` and display a table of active tasks.
* `pm show <id>`: Render the Markdown of a specific ticket in the terminal.

### Phase 2: State & Metadata Management

* `pm move <id> <status>`: Update the YAML `status` field (validated against `workflow.yaml`).
* `pm comment <id> "message"`: Append a timestamped comment to the "Discussion" section.

### Phase 3: Search & Indexing

* **Local Cache:** To avoid slow filesystem crawls, implement a local SQLite indexer that updates whenever a `pm` command is run.

---

## 5. Instructions for the Building LLM

1. **Language:** Create a Go project (with `Cobra`) for the CLI.
2. **Validation:** Ensure every "Write" operation validates the YAML against a schema.
3. **Git Integration:** Use a library like `GitPython` or shell out to `git` to automate the `git add` process for ticket changes.
