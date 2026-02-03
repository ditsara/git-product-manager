Here is a comprehensive technical specification for Git Product Manager, a
Git-based project management system. This document is designed to be
"LLM-ready," providing the necessary structure, logic, and context for an AI Agent.

---

# Specification: Git Product Manager (GPM)

## 1. Vision & Core Philosophy

The goal is to eliminate the context switch between coding and project management. All "tickets" are stored as structured data (YAML + Markdown) within the same repository as the code.

* **Single Source of Truth:** The repository contains both the "what" (tasks) and the "how" (code).
* **GitOps Workflow:** CRUD operations are performed via a CLI, and changes are committed/pushed like code.
* **Process as Code:** Workflows, labels, and validation rules are defined in version-controlled config files.
* **Auditability:** Git history serves as the immutable audit trail for ticket lifecycles.

---

## 1.5. Development Workflow (LLM Collaboration)

This project uses an **iterative, ticket-driven development workflow** designed for LLM-human collaboration. Future AI agents working on this codebase should follow this process:

### The Workflow

1. **Idea/Request:** User provides a feature request, bug report, or improvement idea
   - May be informal ("we should fix X") or detailed
   - May reference existing code or be completely new

2. **Ticket Creation:** LLM creates a comprehensive ticket with:
   - **LLM Attribution:** Add `[Sonnet 4.5]` (or appropriate model identifier) at the start of the description
     - This helps future users know the ticket was AI-generated
     - Update with your model name/version
   - **Problem Statement:** Clear description of the issue or feature
   - **Solution Approach:** Recommended implementation strategy
   - **Edge Cases:** Explicitly defined behavior for corner cases
   - **Implementation Steps:** Actionable checklist items
   - **Acceptance Criteria:** Clear, testable outcomes
   - **Code Examples:** When helpful, include example code or expected behavior
   - Use `pm new` command to create the ticket file
   - Edit the markdown file directly to add full details

3. **Review & Refinement:** User reviews the ticket
   - User may request clarifications
   - User may challenge assumptions or suggest alternatives
   - User may ask for edge cases to be explicitly defined
   - LLM updates the ticket based on feedback
   - Iterate until both parties are satisfied with the specification

4. **Implementation:** LLM implements the ticket
   - Follow the implementation steps checklist
   - Check off items (`- [x]`) as they're completed
   - Write tests for edge cases identified in the ticket
   - Run tests frequently to ensure nothing breaks
   - Update the ticket if implementation reveals new considerations

5. **Verification & Completion:** 
   - Verify all acceptance criteria are met
   - Check off all acceptance criteria (`- [x]`)
   - Mark ticket as done: `pm move TICKET-ID done`
   - Add notes about what was done differently than planned (if applicable)

### Key Principles

**Tickets are the specification:**
- Don't start implementing before the ticket is well-defined
- When in doubt, ask for clarification rather than making assumptions
- Tickets should be detailed enough that another LLM could implement them

**Be explicit about edge cases:**
- User may ask: "what happens if X?"
- Document the answer in the ticket before implementing
- Edge cases often become test cases

**Replacement over append (example from GPM-4):**
- When making design decisions, document the rationale
- Example: "Array updates REPLACE existing values (not append) for consistency with single-value fields"

**Checklists drive implementation:**
- Implementation Steps: What to build
- Acceptance Criteria: How to verify it works
- Check items off as you complete them
- If you deviate from the plan, add a note explaining what you did instead

### Example Exchange

```
User: "when the title is long, it makes the columns not line up in pm list"

LLM: [Creates GPM-12 ticket with:
  - Problem description with example
  - Three solution options (simple truncation, terminal-aware, table library)
  - Recommendation: Option 1 for simplicity
  - Edge cases: Unicode, emoji, empty strings
  - Implementation steps
  - Acceptance criteria]

User: "okay let's implement GPM-12. note that we will choose Option 1. 
       Create a separate enhancement ticket for Option 3."

LLM: [Updates GPM-12 to focus on Option 1
     Creates GPM-15 for Option 3 (table library)
     Implements GPM-12:
       - Adds truncate() function
       - Updates list.go to use it
       - Checks off implementation steps
       - Verifies acceptance criteria
       - Marks ticket as done]
```

### Benefits of This Workflow

- **Reduces rework:** Specification agreed upon before coding
- **Knowledge transfer:** Tickets document decisions for future maintainers
- **Testability:** Edge cases identified upfront become test cases
- **Transparency:** User can see what you're planning before you build it
- **Auditability:** Git history + ticket history = complete project narrative

### Using GPM to Build GPM

This project uses GPM to manage its own development (dogfooding). All tickets are in `.pm/tickets/`. Use `pm list` to see current work, `pm show TICKET-ID` to read specifications, and create new tickets for any work you're doing.

---

## 2. File Architecture

### 2.1 Directory Structure

To keep the root clean, all data resides in a hidden `.pm/` directory at the repository root.

```text
project-root/
├── .pm/
│   ├── tickets/              # Directory of all tasks/stories
│   │   ├── PROJ-a1b2c3.md    # Ticket file
│   │   ├── PROJ-a1b2c3/      # Comments for PROJ-a1b2c3
│   │   │   ├── 2026-01-31T09-00-00Z-alice.md
│   │   │   └── 2026-02-01T14-30-00Z-bob.md
│   │   ├── PROJ-d4e5f6.md
│   │   └── PROJ-d4e5f6/
│   ├── config/
│   │   ├── workflow.yaml     # State definitions
│   │   ├── labels.yaml       # Allowed tags
│   │   └── templates/        # Ticket templates
│   │       ├── story.md
│   │       ├── task.md
│   │       ├── bug.md
│   │       └── epic.md
│   ├── .gitignore            # Ignore cache files
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
parent: ARCH-45      # Parent ticket (epic or story)
depends_on: []       # Blocked by these tickets
blocks: []           # This ticket blocks these tickets
related: []          # Related tickets (duplicates, see-also)
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

```

### 2.3 Comment Schema (Separate Files)

Comments are stored as individual markdown files in a directory named after the ticket ID. This prevents merge conflicts when multiple people comment simultaneously.

**Directory:** `.pm/tickets/{TICKET-ID}/`

**Filename Format:** `{ISO8601-timestamp}-{author}.md`
- Timestamp uses hyphens instead of colons for filesystem compatibility
- Example: `2026-02-01T14-30-00Z-alice.md`

**Comment File Content:**

```markdown
---
author: alice
timestamp: 2026-02-01T14:30:00Z
---

Should we use OAuth2 library or implement from scratch? I'm concerned about security implications.
```

**Benefits:**
- **No merge conflicts:** Each comment is an atomic file
- **Clean git history:** Each comment is its own commit
- **Scalable:** Tickets with many comments don't bloat the main file
- **Flexible:** Can delete/edit individual comments if needed

---

## 3. Technical Requirements

### 3.1 Uniqueness & Identification

* **K-Sorted IDs (KSUID):** To prevent merge conflicts in a distributed environment, use a configurable prefix + KSUID suffix.
  * **Format:** `{PREFIX}-{base62(timestamp + random)}` (e.g., `PROJ-a1b2c3d4`)
  * **Implementation:** Use the KSUID algorithm (lexicographically sortable, ~27 chars encoded)
  * **Prefix:** Defined in `.pm/config/project.yaml` (default: repository name uppercase)
  * **Collision Resistance:** 160-bit random component ensures uniqueness across branches
  * **Benefits:** Natural chronological sorting, no coordination needed between developers

### 3.2 Workflow Configuration (`workflow.yaml`)

States are simple labels that tickets can have. There are no enforced transitions - teams can move tickets between any states freely. Git history provides the audit trail of state changes.

```yaml
states:
  - backlog
  - todo
  - in-progress
  - done

initial_state: backlog

# Optional: Semantic groupings for filtering and reporting
state_groups:
  active: [todo, in-progress]
  completed: [done]
  incomplete: [backlog, todo, in-progress]
```

**Validation Rules:**
- `status` field must be one of the defined states
- `initial_state` must exist in the `states` list
- State names must be lowercase with hyphens (e.g., `in-progress`, not `In Progress`)

### 3.3 The "Bad YAML" Guardrail

The implementation must include a **Validation Layer**:

* **Auto-Fix (Optional, Configurable):** On save, optionally normalize formatting:
  * Fix indentation (2 spaces)
  * Trim trailing whitespace
  * Ensure newline at end of file
  * Log all auto-fixes to stderr
  * Disabled by default; enable via `pm config set auto-fix true`

* **Linter (`pm validate`):** Run validation on demand or before commits:
  * **Required Fields:** `id`, `title`, `type`, `status`, `created_at`
  * **Valid Enums:** `type` ∈ {epic, story, task, bug}, `status` must be in configured states list
  * **Date Format:** ISO8601 with UTC timezone (e.g., `2026-01-31T09:00:00Z`)
  * **ID Format:** Must match `{PREFIX}-[a-zA-Z0-9]+`
  * **YAML Syntax:** Must parse without errors
  * **Reference Integrity:** All ticket IDs in `parent`, `depends_on`, `blocks`, and `related` must exist
  * **Relationship Arrays:** Must be valid YAML arrays (empty arrays allowed)

* **Git Hook (Optional Setup):** `pm init` offers to install a pre-commit hook:
  ```bash
  #!/bin/bash
  pm validate || exit 1
  ```

### 3.4 Timestamp Handling

* **All timestamps use UTC:** ISO8601 format with `Z` suffix
* **Auto-update:** The CLI automatically updates `updated_at` on any modification
* **Immutable:** `created_at` is set once and never changes
* **Comment timestamps:** Appended in the same format

### 3.5 Ticket Relationships

The system supports a **computed relationship model** where relationships are stored in ticket metadata and indexed in the SQLite cache for fast queries.

#### Relationship Types

1. **Parent-Child Hierarchy (`parent` field)**
   * Generic parent reference supporting flexible hierarchies:
     - Epic → Story
     - Epic → Task  
     - Epic → Bug
     - Story → Task
     - Story → Bug
   * Stored as: `parent: PARENT-ID` (single value, optional)
   * Query children: `pm list --parent EPIC-123`
   * Visualize tree: `pm tree EPIC-123`

2. **Dependencies (`depends_on` and `blocks` fields)**
   * `depends_on`: Array of ticket IDs this ticket cannot start without
   * `blocks`: Array of ticket IDs that cannot start until this is done
   * Stored as: `depends_on: [TICKET-1, TICKET-2]`
   * Used for scheduling and detecting blockers

3. **Related Tickets (`related` field)**
   * Loose associations (duplicates, see-also, related work)
   * Stored as: `related: [TICKET-X, TICKET-Y]`
   * Bidirectional linking not enforced

#### Validation Rules

* **No Circular Dependencies:** The validator detects cycles in `depends_on` relationships
* **Reference Integrity:** All referenced ticket IDs must exist
* **No Self-Reference:** A ticket cannot reference itself in any relationship field
* **User Responsibility:** The system does NOT enforce semantic rules (e.g., stories containing epics). Users must maintain logical hierarchies.

#### SQLite Cache Schema

The cache builds a relationship graph and indexes comments for fast queries:

```sql
-- Ticket metadata table
CREATE TABLE tickets (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  type TEXT NOT NULL,
  status TEXT NOT NULL,
  priority TEXT,
  assignee TEXT,
  parent TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  body TEXT  -- Full markdown content for FTS
);

CREATE VIRTUAL TABLE tickets_fts USING fts5(
  id, title, body,
  content='tickets'
);

-- Relationship graph
CREATE TABLE relationships (
  from_ticket TEXT NOT NULL,
  to_ticket TEXT NOT NULL,
  relationship_type TEXT NOT NULL, -- 'parent', 'depends_on', 'blocks', 'related'
  PRIMARY KEY (from_ticket, to_ticket, relationship_type)
);

CREATE INDEX idx_from ON relationships(from_ticket);
CREATE INDEX idx_to ON relationships(to_ticket);

-- Comment index
CREATE TABLE comments (
  ticket_id TEXT NOT NULL,
  author TEXT NOT NULL,
  timestamp TEXT NOT NULL,
  filepath TEXT NOT NULL,  -- Relative path from .pm/tickets/
  PRIMARY KEY (ticket_id, timestamp, author)
);

CREATE INDEX idx_ticket_comments ON comments(ticket_id, timestamp);
```

---

## 4. Feature Set for Prototype

### Phase 1: The Core CLI

* `pm init [PATH] [--prefix PREFIX]`: Setup the `.pm` directory and default configs
  * **PATH:** Optional directory path (defaults to current directory `.`)
  * Creates `.pm/` directory structure at specified location
  * Generates default configuration files:
    - `.pm/config/workflow.yaml` with standard states (backlog, todo, in-progress, done)
    - `.pm/config/labels.yaml` with common labels (backend, frontend, bug, feature, etc.)
    - `.pm/config/templates/` with story, task, bug, and epic templates
  * Creates `.pm/.gitignore` to exclude `.cache.db` and temporary files
  * Initializes empty `.pm/tickets/` directory
  * Initializes SQLite cache at `.pm/.cache.db`
  * Optionally installs pre-commit hook (interactive prompt: "Install git hook? [y/N]")
  * Sets project prefix in config (default: uppercase of repository/directory name)
  * **Exit conditions:**
    - Error if `.pm/` already exists (prevent accidental overwrite)
    - Suggest `pm init --force` to reinitialize
  * **Output:** Confirmation message with next steps:
    ```
    ✓ Initialized .pm directory
    ✓ Created default workflow with 4 states
    ✓ Created 15 default labels
    ✓ Created 4 ticket templates
    
    Next steps:
      pm new "Your first ticket"
      pm list
    ```

* `pm new [--type TYPE] "Title"`: Generate a new ticket file from template
  * **Type:** story (default), task, bug, epic
  * Generate KSUID-based ID
  * Populate template with defaults (status=backlog, created_at=now)
  * Open in `$EDITOR` if set, otherwise save directly
  * Auto-add to git staging area

* `pm list [--status STATUS] [--assignee USER] [--label TAG] [--parent ID]`: Display filtered table
  * Query SQLite cache for performance
  * Columns: ID, Title, Type, Status, Assignee, Priority
  * Support multiple filters (AND logic)
  * `--parent ID`: Show all direct children of a ticket
  * Sort by updated_at descending
  * Color-coded output (configurable)

* `pm show <id> [--no-comments]`: Render ticket with syntax highlighting
  * Parse YAML front-matter and display as formatted table
  * Render Markdown body with terminal formatting
  * Show relationship information (parent, children count, dependencies)
  * **Display comments:** Read all files from `.pm/tickets/{id}/` directory
    - Sort chronologically by timestamp
    - Format: `@{author} ({timestamp}): {comment_text}`
    - Skip if `--no-comments` flag provided
  * Show git history summary (author, date of last 3 changes)

* `pm history <id>`: Display state change audit trail
  * Parse git history for the ticket file
  * Extract and display status field changes over time
  * Format: `YYYY-MM-DD HH:MM  author  old_state → new_state`
  * Show commit messages for context

### Phase 2: State & Metadata Management

* `pm move <id> <status> [--no-commit]`: Update the YAML `status` field
  * Validate status exists in `workflow.yaml` states list
  * Update `updated_at` timestamp
  * Auto-commit with message: `chore(pm): Move {id} to {status}`
  * Skip commit if `--no-commit` flag provided

* `pm edit <id> [--field FIELD]`: Modify ticket metadata or content
  * **No arguments:** Opens the full ticket file in `$EDITOR`
  * **With --field:** Directly update specific field (e.g., `--field assignee=alice`)
    - Supported fields: `assignee`, `priority`, `points`, `labels`, `parent`
    - Array fields (labels, depends_on, blocks, related) accept comma-separated values
  * **Editor selection (standard fallback chain):**
    - `$VISUAL` (for full-screen editors)
    - `$EDITOR` (for line editors)
    - `editor` command (Debian/Ubuntu alternatives system)
    - `nano` (common modern default)
    - `vi` (POSIX fallback, always present)
  * **Wait behavior:** Waits for editor to close before continuing
  * **Post-edit validation:**
    - Parse and validate YAML after editor closes
    - If invalid, prompt to re-edit or discard changes
  * **Auto-update:** Sets `updated_at` timestamp
  * **Auto-commit:** Commit with message: `chore(pm): Edit {id}`
  * **Examples:**
    ```bash
    pm edit PROJ-123                    # Open in editor
    pm edit PROJ-123 --field priority=high
    pm edit PROJ-123 --field labels=bug,critical
    ```

* `pm comment <id> [-m MESSAGE] [--author NAME]`: Add a comment to a ticket
  * **Interactive mode (default):** Opens editor for comment composition
    - Uses standard fallback chain: `$VISUAL` → `$EDITOR` → `editor` → `nano` → `vi`
    - Pre-populates with template showing ticket context (similar to `git commit`)
    - Comment lines starting with `#` are ignored
    - Aborts if editor exits with empty content
  * **Direct mode (`-m`):** Provide comment message directly
    - `pm comment PROJ-123 -m "Looks good to merge"`
    - Skips editor, creates comment immediately
  * **Comment file creation:**
    - Creates directory `.pm/tickets/{id}/` if it doesn't exist
    - Generates filename: `{ISO8601-timestamp}-{author}.md`
    - Timestamp uses hyphens for filesystem compatibility (e.g., `2026-02-01T14-30-00Z`)
    - Author defaults to git config `user.name` or can be overridden with `--author`
  * **Comment content:**
    - YAML front-matter with `author` and `timestamp`
    - Markdown body with comment text
  * **Auto-commit:** Commit with message: `comment(pm): Add comment to {id}`
  * **Cache update:** Add entry to `comments` table in SQLite cache
  * **Examples:**
    ```bash
    pm comment PROJ-123                          # Open editor
    pm comment PROJ-123 -m "LGTM"                # Direct comment
    pm comment PROJ-123 -m "Fixed" --author bob  # Override author
    ```

* `pm assign <id> <user>`: Update assignee field
  * Shorthand for `pm edit <id> --field assignee=<user>`

### Phase 3: Search & Indexing

* **SQLite Cache (`.pm/.cache.db`):**
  * **Schema:** Mirrors ticket YAML fields with FTS (Full-Text Search) on title/body
  * **Update Strategy:** Incremental updates based on file mtime
  * **Rebuild:** `pm reindex` to force full rebuild from filesystem
  * **Ignored by Git:** Listed in `.pm/.gitignore`

* **Search Commands:**
  * `pm search <query>`: Full-text search across title and markdown body
  * `pm list --label <tag>`: Filter by label (supports multiple: `--label auth --label api`)
  * `pm list --parent <id>`: Show all tickets with the specified parent

### Phase 4: Relationship Management

* `pm link <id> <target-id> [--type TYPE]`: Create a relationship between tickets
  * **Types:** `parent`, `depends-on`, `blocks`, `related`
  * Default: `related`
  * Updates the appropriate field in the source ticket
  * Auto-commit with message: `chore(pm): Link {id} to {target-id} ({type})`

* `pm unlink <id> <target-id> [--type TYPE]`: Remove a relationship
  * Removes the target from the relationship array
  * If `--type` not specified, removes from all relationship fields

* `pm tree <id> [--depth N]`: Visualize ticket hierarchy as ASCII tree
  * Shows parent-child relationships recursively
  * Default depth: unlimited
  * Example output:
    ```
    EPIC-123: Implement Authentication
    ├── STORY-456: OAuth2 Login
    │   ├── TASK-789: Setup Google Provider
    │   └── TASK-790: Create JWT Middleware
    └── STORY-457: Password Reset Flow
        └── TASK-791: Email Template
    ```

* `pm blocked [<id>]`: Show dependency information
  * No ID: List all tickets with unresolved dependencies
  * With ID: Show what blocks this ticket and what it blocks

### Phase 5: Advanced Features (Future)

* `pm validate`: Run linter on all tickets (Phase 1 dependency)
* `pm stats`: Show metrics (tickets by status, burndown data)
* `pm graph <id> --format dot`: Generate GraphViz dependency graph
* `pm sync`: Push/pull workflow with conflict detection
* bash completion support
* customize the `.pm` directory, e.g., `pm --root .other-pm-dir`

---

## 5. Implementation Guidelines

### 5.1 Technology Stack

* **Language:** Go 1.21+
* **CLI Framework:** [Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper) for config
* **YAML Parsing:** `gopkg.in/yaml.v3` with strict mode
* **Git Operations:** `go-git/go-git/v5` (pure Go) or shell out to `git` CLI
* **SQLite:** `modernc.org/sqlite` (CGo-free) or `mattn/go-sqlite3`
* **Markdown Rendering:** `charmbracelet/glamour` for terminal output
* **ID Generation:** `segmentio/ksuid` library
* **Database Migrations:** `golang-migrate/migrate` for schema versioning

### 5.2 Project Structure

```
cmd/
  pm/
    main.go
    init.go
    new.go
    list.go
    show.go
    move.go
    comment.go
    ...
internal/
  ticket/
    ticket.go       # Ticket struct and parsing
    validator.go    # Validation logic
    comment.go      # Comment file operations
  config/
    workflow.go     # Workflow configuration
    labels.go
  cache/
    sqlite.go       # Cache operations
    migrate.go      # Migration runner
  git/
    operations.go   # Git add/commit helpers
migrations/
  000001_initial_schema.up.sql
  000001_initial_schema.down.sql
  000002_add_comments_table.up.sql
  000002_add_comments_table.down.sql
scripts/
  build.sh          # Build binary to bin/ directory
  test-local.sh     # Test commands in sandbox/
bin/
  pm                # Compiled binary (gitignored)
sandbox/
  .pm/              # Test project directory (gitignored)
go.mod
go.sum
.gitignore
README.md
AGENTS.md          # This specification file
```

### 5.3 Build and Development Workflow

**Default Build Location:** `bin/pm`
- Go's default: `go build` creates binary in current directory
- This project uses `bin/` directory for compiled binaries (gitignored)

**Build Script:** `scripts/build.sh`
```bash
#!/bin/bash
set -e

echo "Building pm..."
mkdir -p bin
go build -o bin/pm ./cmd/pm

echo "✓ Built binary: bin/pm"
echo ""
echo "To use:"
echo "  ./bin/pm init sandbox"
echo "  ./bin/pm --help"
```

**Local Testing Script:** `scripts/test-local.sh`
```bash
#!/bin/bash
set -e

# Build the binary
./scripts/build.sh

# Create a clean test environment
echo ""
echo "Setting up test environment in sandbox/..."
rm -rf sandbox
mkdir -p sandbox

cd sandbox

# Initialize and test
echo ""
echo "Running: pm init ."
../bin/pm init .

echo ""
echo "Running: pm new 'Test ticket'"
../bin/pm new "Test ticket"

echo ""
echo "Running: pm list"
../bin/pm list

echo ""
echo "✓ Local test complete"
echo ""
echo "Test directory: sandbox/"
echo "To continue testing:"
echo "  cd sandbox"
echo "  ../bin/pm <command>"
```

**Root `.gitignore` additions:**
```gitignore
# Build artifacts
/bin/
/dist/

# Local testing
/sandbox/

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
```

**Development Commands:**
```bash
# Build binary
./scripts/build.sh

# Or directly with go
go build -o bin/pm ./cmd/pm

# Run tests
go test ./...

# Test locally in sandbox
./scripts/test-local.sh

# Or manually
./bin/pm init sandbox
cd sandbox && ../bin/pm new "My ticket"

# Install to $GOPATH/bin for system-wide use
go install ./cmd/pm
```

**Makefile (optional but recommended):**
```makefile
.PHONY: build test install clean

build:
	@./scripts/build.sh

test:
	@go test ./...

test-local: build
	@./scripts/test-local.sh

install:
	@go install ./cmd/pm

clean:
	@rm -rf bin/ sandbox/ dist/
	@echo "✓ Cleaned build artifacts"

all: clean build test
```
  000001_initial_schema.up.sql
all: clean build test
```

### 5.4 Database Migrations

Use **golang-migrate/migrate** for managing SQLite schema versions.

**Migration Directory:** `migrations/` at project root

**Naming Convention:** `{version}_{description}.{up|down}.sql`
- **Version:** 6-digit zero-padded number (000001, 000002, etc.)
- **Description:** Snake_case description of the change
- **Direction:** `up.sql` for applying, `down.sql` for rollback

**Example Migration Pair:**

`migrations/000001_initial_schema.up.sql`:
```sql
CREATE TABLE schema_migrations (
  version INTEGER PRIMARY KEY,
  dirty BOOLEAN NOT NULL
);

CREATE TABLE tickets (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  type TEXT NOT NULL,
  status TEXT NOT NULL,
  priority TEXT,
  assignee TEXT,
  parent TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  body TEXT
);

CREATE VIRTUAL TABLE tickets_fts USING fts5(
  id, title, body,
  content='tickets'
);

CREATE TABLE relationships (
  from_ticket TEXT NOT NULL,
  to_ticket TEXT NOT NULL,
  relationship_type TEXT NOT NULL,
  PRIMARY KEY (from_ticket, to_ticket, relationship_type)
);

CREATE INDEX idx_from ON relationships(from_ticket);
CREATE INDEX idx_to ON relationships(to_ticket);
```

`migrations/000001_initial_schema.down.sql`:
```sql
DROP TABLE IF EXISTS relationships;
DROP TABLE IF EXISTS tickets_fts;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS schema_migrations;
```

**Integration:**
- **On `pm init`:** Run all migrations to create fresh cache database
- **On any command:** Check schema version, auto-migrate if needed
- **Migration tracking:** Uses `schema_migrations` table (created by golang-migrate)

**Implementation:**
```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/sqlite3"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dbPath string) error {
    m, err := migrate.New(
        "file://migrations",
        "sqlite3://"+dbPath,
    )
    if err != nil {
        return err
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

### 5.5 Error Handling

* All validation errors must include:
  * Ticket ID (if applicable)
  * Field name causing the error
  * Expected vs. actual value
  * Suggestion for fix

* Example: `Error: Invalid status 'completed' for PROJ-abc123. Valid states: [backlog, todo, in-progress, done]. Did you mean 'done'?`

### 5.6 Concurrency & Merge Conflicts

* **Detection:** Before `pm` operations, check if `.pm/tickets/` has uncommitted changes
* **Git History Audit:** All state changes are tracked via git commits, providing immutable audit trail
* **Conflict Resolution:** If git merge conflict detected in a ticket file:
  * Parse both versions
  * If only `updated_at` differs, take the latest
  * If metadata conflicts, prompt user to resolve manually
  * Preserve all comments from both versions
* **State Change History:** Use `pm history <id>` to view who changed states and when
* **Comment Conflicts:** Individual comment files prevent conflicts; multiple people can comment simultaneously

### 5.7 Configuration Files

* **Location:** `.pm/config/`
* **Validation on Load:** All YAML configs validated at CLI startup
* **User Overrides:** Support `~/.pm/config.yaml` for global defaults

### 5.8 Testing Requirements

* Unit tests for:
  * Ticket parsing and validation
  * State validation (must be in configured list)
  * KSUID generation uniqueness
  * Cache synchronization
  * Git history parsing for state changes
  * Comment file creation and parsing
  * Database migrations (up/down)
* Integration tests:
  * Full CLI workflows (init → new → move → comment)
  * Git operations (commit, merge simulation)
  * State change history extraction
  * Migration rollback scenarios
* Use `t.TempDir()` for isolated filesystem tests

---

## 6. Build Stages

This project will be built in three distinct stages, each resulting in a viable, testable product.

### Stage 1: Core Ticket Management (MVP)

This stage focuses on the essential functionality for a single user to manage tickets locally. It establishes the core data structures and commands but omits collaborative features like comments and complex relationships.

- [x] **`pm init`**: Initialize the `.pm` directory, configs, and an initial SQLite database (without comment/relationship tables).
- [x] **`pm new`**: Create new tickets from templates.
- [x] **`pm list`**: List all tickets with basic filtering (status, type).
- [x] **`pm show`**: Display a single ticket's content (no comments).
- [x] **`pm edit`**: Open a ticket in `$EDITOR` and update basic fields.
- [x] **`pm move`**: Change a ticket's status.
- [x] **Validation**: Implement the "Bad YAML" guardrail for all ticket write operations.
- [x] **Database**:
    - [x] Implement the `tickets` table in SQLite.
    - [x] Set up `golang-migrate` with the initial schema for the `tickets` table only.
- [x] **Tests**:
    - [x] Unit tests for ticket parsing and validation.
    - [x] Integration tests for the `init` -> `new` -> `list` -> `show` -> `edit` -> `move` workflow.

### Stage 1.5: Refinements and Usability Improvements

This stage refines the MVP based on initial usage feedback, focusing on making the tool more user-friendly and predictable.

- [x] **`pm init` improvements**:
    - [x] Make `--prefix` flag required (no default)
    - [x] Convert prefix to uppercase automatically (e.g., `ticket` → `TICKET`)
    - [x] Verify SQLite database is actually created and accessible after migration
    - [x] Update error messages to guide users on prefix requirements
- [x] **Ticket ID format change**:
    - [x] Replace KSUID-based IDs with sequential integers (`PREFIX-1`, `PREFIX-2`, etc.)
    - [x] Implement filesystem-based ID generation: scan `.pm/tickets/` to find highest number
    - [x] Handle gaps in sequence gracefully (e.g., 1,2,4,5 → next is 6)
    - [x] Do NOT rely on SQLite cache for ID generation (support manual ticket creation)
    - [x] Update ID validation regex to accept `{PREFIX}-\d+` format
- [x] **Update `pm new` command**:
    - [x] Read prefix from `.pm/config/project.yaml` (created during init)
    - [x] Generate next sequential ID by scanning existing tickets
    - [x] Update templates to work with new ID format
- [x] **Configuration file**:
    - [x] Create `.pm/config/project.yaml` during init with structure:
      ```yaml
      prefix: TICKET  # User-specified, uppercased
      ```
    - [x] Add config loading utility in `internal/config/project.go`
- [x] **Tests**:
    - [x] Unit tests for sequential ID generation with gaps
    - [x] Integration test: init with custom prefix, create multiple tickets
    - [x] Test uppercase conversion of prefix
    - [x] Verify database initialization completes successfully

**Rationale:**
- **Required prefix**: Forces intentional naming, prevents accidental defaults
- **Uppercase convention**: Standard practice for ticket systems (JIRA, etc.)
- **Sequential IDs**: Much more memorable and user-friendly than KSUIDs
- **Filesystem-first**: Ensures the system works even if cache is corrupted/deleted

### Stage 1.6: UX Polish and Cache Strategy

This stage improves the user experience with better help messages and establishes a clear strategy for SQLite cache synchronization to handle manual ticket edits.

**Key deliverables:**
- Better help messages when commands are run without arguments ([GPM-16](.pm/tickets/GPM-16.md))
- Lazy migration check to auto-recover from missing/outdated cache ([GPM-10](.pm/tickets/GPM-10.md))
- Cache metadata table with automatic staleness detection and sync ([GPM-17](.pm/tickets/GPM-17.md))

**Rationale:**
- **Lazy sync**: Balances performance with correctness—cache is fast but never shows stale data
- **Transparent to users**: Most users won't need to think about the cache; it "just works"
- **Better UX**: New users discover commands naturally without reading docs

### Stage 2: Collaboration and History

**See ticket: [GPM-1](.pm/tickets/GPM-1.md)**

This stage introduces features for team collaboration and auditing, centered around the conflict-free comment system and git history analysis.

**Key deliverables:**
- `pm comment` - Full comment system with separate files
- `pm show` - Enhanced to display comments
- `pm history` - State change auditing via git history
- `pm assign` - Assignee shorthand command
- Database migration for comments table
- Comprehensive test coverage

### Stage 3: Advanced Relationships and Search

**See ticket: [GPM-2](.pm/tickets/GPM-2.md)**

This final stage completes the vision by adding powerful relationship tracking, visualization, and efficient searching capabilities.

**Key deliverables:**
- `pm link` & `pm unlink` - Relationship management
- `pm tree` - Hierarchy visualization
- `pm blocked` - Dependency tracking
- `pm search` - Full-text search
- Enhanced `pm list` with parent filtering
- Reference integrity validation
- Database migrations for relationships and FTS tables
- Comprehensive test coverage

---

## 7. Future Improvements

### Enhanced Cache Control (Future)

While Stage 1.6 implements automatic lazy sync for cache correctness, future versions could add explicit cache management commands for power users and debugging:

- **`pm sync` command**: Manual cache rebuild from filesystem
  - Useful for troubleshooting
  - Forces full rescan regardless of timestamps
  - Displays progress for large repositories
  
- **`--no-cache` flag**: Bypass cache entirely on read operations
  - Example: `pm list --no-cache`
  - Reads directly from filesystem
  - Useful for debugging cache issues
  - Performance trade-off acceptable for debugging scenarios
  
- **Staleness warnings**: Display notification if cache hasn't synced in >5 minutes
  - Example: "⚠ Cache may be outdated. Run `pm sync` to refresh."
  - Helps users understand when manual sync might be helpful
  
- **Cache statistics**: `pm cache-info` command to show cache health
  - Last sync timestamp
  - Number of cached tickets
  - Cache file size
  - Staleness status

**Rationale for deferring:**
- Lazy sync (Stage 1.6) handles 99% of use cases automatically
- Additional commands add complexity for diminishing returns
- Can be added later based on actual user needs and feedback
- Keeps the core tool simple and focused

### Bash Completion (Future)

Add shell completion support for better UX with command-line interaction:

- **`pm completion` command**: Generate completion script for bash/zsh/fish
  - Leverages Cobra's built-in completion generation
  - Example: `pm completion bash > /etc/bash_completion.d/pm`
  - Instructions: `source <(pm completion bash)` for immediate use
  
- **Auto-complete features**:
  - Subcommands: `pm li<TAB>` → `pm list`
  - Flags: `pm new --ty<TAB>` → `pm new --type`
  - Flag values: `pm new --type t<TAB>` → `pm new --type task`
  - Ticket IDs: `pm show TEST-<TAB>` → shows available ticket IDs
  - Status values: `pm move TICKET-1 <TAB>` → shows valid states from workflow.yaml
  
- **Implementation**:
  - Add completion command using `rootCmd.GenBashCompletion()`
  - Register custom completions for ticket IDs (read from `.pm/tickets/`)
  - Register custom completions for statuses (read from `workflow.yaml`)
  - Generate man pages alongside completion scripts

**Rationale for deferring:**
- Not critical for core functionality
- Easy to add later (~15 minutes with Cobra)
- Command structure should stabilize first
- Better to add when there are real users requesting it
