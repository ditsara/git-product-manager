# SQLite Cache Data Model

## Overview

The Git Product Manager (GPM) uses a local SQLite database (`.pm/.cache.db`) to
provide fast querying and filtering of tickets without parsing YAML files on
every command. The cache is **read-only from the user's perspective** and is
automatically synchronized with the filesystem using a lazy sync strategy.

**Key Principles:**
- **Single Source of Truth:** Ticket markdown files in `.pm/tickets/` are authoritative
- **Lazy Synchronization:** Cache auto-syncs when stale (detected via filesystem mtimes)
- **Transactional Updates:** All cache operations use transactions for consistency
- **Bulk Operations:** Syncs use bulk inserts for performance

---

## Schema Versions

The cache uses `golang-migrate` for version control:

| Version | Migration | Description |
|---------|-----------|-------------|
| 1 | `000001_initial_schema` | Core `tickets` table |
| 2 | `000002_add_cache_metadata` | Sync timestamp tracking |
| 3 | `000003_add_comments_table` | Comment indexing |
| 4 | `000004_add_relationships_table` | Dependency graph |

---

## Tables

### 1. `tickets`

**Purpose:** Primary index of all ticket metadata for fast filtering and searching.

**Schema:**
```sql
CREATE TABLE tickets (
  id TEXT PRIMARY KEY,           -- Ticket ID (e.g., "PROJ-123")
  title TEXT NOT NULL,           -- Ticket title
  type TEXT NOT NULL,            -- Enum: epic, story, task, bug
  status TEXT NOT NULL,          -- Current workflow state
  priority TEXT,                 -- Optional: high, medium, low
  assignee TEXT,                 -- Optional: username
  parent TEXT,                   -- Optional: parent ticket ID
  created_at TEXT NOT NULL,      -- ISO8601 UTC timestamp
  updated_at TEXT NOT NULL,      -- ISO8601 UTC timestamp
  body TEXT                      -- Full markdown content (for future FTS)
);
```

**Columns:**
- **`id`**: Unique identifier (`{PREFIX}-{NUMBER}` format, e.g., `PROJ-1`)
- **`title`**: Short description extracted from YAML front-matter
- **`type`**: Ticket classification (epic, story, task, bug)
- **`status`**: Current state from workflow.yaml
- **`priority`**: Optional prioritization (high, medium, low, or custom)
- **`assignee`**: Username from git config or manually set
- **`parent`**: Reference to parent ticket (supports hierarchical organization)
- **`created_at`**: Immutable creation timestamp
- **`updated_at`**: Last modification timestamp (auto-updated)
- **`body`**: Full markdown content after YAML front-matter (reserved for future full-text search)

**Relationships:**
- **Self-referential:** `parent` field references another `tickets.id`
- **Comments:** One-to-many with `comments` table (via `ticket_id`)
- **Dependencies:** Many-to-many with itself via `relationships` table

**Usage:**
- Powers `pm list` with filtering by status, assignee, type, parent
- Enables fast lookups in `pm show` and `pm edit`
- Foundation for future search and reporting features

---

### 2. `cache_metadata`

**Purpose:** Stores cache synchronization state and configuration.

**Schema:**
```sql
CREATE TABLE cache_metadata (
  key TEXT PRIMARY KEY,          -- Metadata key (e.g., "last_sync_timestamp")
  value TEXT NOT NULL            -- Metadata value (string format)
);
```

**Current Keys:**
- **`last_sync_timestamp`**: ISO8601 UTC timestamp of last successful sync
  - Example: `"2026-02-15T03:00:00Z"`
  - Used by `ShouldSync()` to detect stale cache
  - Updated atomically with ticket data in transactions

**Usage:**
- **Lazy Sync Detection:** Before read operations, `ShouldSync()` compares this timestamp against ticket file mtimes
- **Initialization:** Set to `1970-01-01T00:00:00Z` on first run to force initial sync
- **Future Extensions:** Can store schema version, feature flags, or user preferences

**Synchronization Logic:**
```go
// Pseudo-code from internal/cache/sync.go
if anyTicketFile.ModTime >= cache_metadata.last_sync_timestamp {
    SyncCache() // Full rebuild
}
```

---

### 3. `comments`

**Purpose:** Index comment files for fast retrieval without filesystem traversal.

**Schema:**
```sql
CREATE TABLE comments (
  ticket_id TEXT NOT NULL,       -- Foreign key to tickets.id
  author TEXT NOT NULL,          -- Comment author username
  timestamp TEXT NOT NULL,       -- ISO8601 UTC timestamp
  filepath TEXT NOT NULL,        -- Relative path from .pm/tickets/
  PRIMARY KEY (ticket_id, timestamp, author)
);

CREATE INDEX idx_ticket_comments ON comments(ticket_id, timestamp);
```

**Columns:**
- **`ticket_id`**: The ticket this comment belongs to (e.g., `PROJ-123`)
- **`author`**: Username from comment YAML front-matter or `--author` flag
- **`timestamp`**: When the comment was created (ISO8601 format)
- **`filepath`**: Relative path to comment file (e.g., `PROJ-123/2026-02-15T03-00-00Z-alice.md`)

**Composite Primary Key:**
- Ensures uniqueness: one comment per (ticket, timestamp, author) tuple
- Supports multiple authors commenting in the same second

**Index:**
- **`idx_ticket_comments`**: Enables efficient chronological ordering when displaying comments

**Relationships:**
- **Many-to-one with `tickets`:** Comments belong to a single ticket (no FK constraint for flexibility)

**Usage:**
- **`pm show TICKET-ID`**: Quickly retrieves all comments for a ticket, sorted chronologically
- **Comment Count:** Can `COUNT(*)` comments per ticket for list views
- **File Loading:** `filepath` column enables direct file reads for comment content

**Note on Storage:**
- Comment **content** is NOT cached (only metadata)
- Actual markdown content loaded from filesystem on demand via `filepath`
- Keeps cache size small and avoids duplication

---

### 4. `relationships`

**Purpose:** Graph representation of ticket dependencies and hierarchies.

**Schema:**
```sql
CREATE TABLE relationships (
  from_ticket TEXT NOT NULL,           -- Source ticket ID
  to_ticket TEXT NOT NULL,             -- Target ticket ID
  relationship_type TEXT NOT NULL,     -- Type: "depends-on" | "blocks"
  PRIMARY KEY (from_ticket, to_ticket, relationship_type)
);

CREATE INDEX idx_from ON relationships(from_ticket);
CREATE INDEX idx_to ON relationships(to_ticket);
CREATE INDEX idx_type ON relationships(relationship_type);
```

**Columns:**
- **`from_ticket`**: Source ticket in the relationship
- **`to_ticket`**: Target ticket in the relationship
- **`relationship_type`**: Type of relationship (currently `depends-on` or `blocks`)

**Relationship Types:**

| Type | Meaning | Ticket Field | Example |
|------|---------|--------------|---------|
| `depends-on` | Source cannot start without target being done | `depends_on: [...]` | PROJ-2 depends-on PROJ-1 |
| `blocks` | Target cannot start until source is done | `blocks: [...]` | PROJ-1 blocks PROJ-2 |

**Composite Primary Key:**
- Prevents duplicate relationships
- Allows same two tickets to have different relationship types

**Indexes:**
- **`idx_from`**: Fast lookup of all dependencies for a ticket ("what does this block?")
- **`idx_to`**: Fast reverse lookup ("what blocks this ticket?")
- **`idx_type`**: Filter relationships by type

**Relationships:**
- **Many-to-many with `tickets`:** Tickets can have multiple dependencies and blockers
- **No FK constraints:** Allows tickets to reference not-yet-created tickets (flexibility)

**Usage:**
- **`pm blocked [TICKET-ID]`**: Find all unresolved dependencies
- **`pm tree TICKET-ID`**: Build hierarchical visualization (uses `parent` field from tickets table)
- **Dependency Graphs:** Export to GraphViz for visual analysis
- **Validation:** Check for circular dependencies before merging

**Derivation from Source:**
- Extracted from ticket YAML arrays (`depends_on`, `blocks`)
- **Example ticket:**
  ```yaml
  depends_on: [PROJ-1, PROJ-5]
  blocks: [PROJ-10]
  ```
  Generates relationships:
  - `(PROJ-X, PROJ-1, depends-on)`
  - `(PROJ-X, PROJ-5, depends-on)`
  - `(PROJ-X, PROJ-10, blocks)`

---

## Entity Relationship Diagram (ERD)

```
┌─────────────────────┐
│   cache_metadata    │
├─────────────────────┤
│ key (PK)            │
│ value               │
└─────────────────────┘

┌─────────────────────────────┐
│         tickets             │
├─────────────────────────────┤
│ id (PK)                     │◄─┐
│ title                       │  │
│ type                        │  │
│ status                      │  │
│ priority                    │  │
│ assignee                    │  │
│ parent (FK) ────────────────┘  │ (self-referential)
│ created_at                  │  │
│ updated_at                  │  │
│ body                        │  │
└─────────────────────────────┘  │
         △                       │
         │                       │
         │ 1                     │
         │                       │
         │                       │
         │ N                     │
┌─────────────────────────────┐  │
│        comments             │  │
├─────────────────────────────┤  │
│ ticket_id (PK, FK)          │──┘
│ timestamp (PK)              │
│ author (PK)                 │
│ filepath                    │
└─────────────────────────────┘
  INDEX: (ticket_id, timestamp)


┌─────────────────────────────┐
│      relationships          │
├─────────────────────────────┤
│ from_ticket (PK, FK) ───────┼───┐
│ to_ticket (PK, FK) ─────────┼───┼──► tickets.id
│ relationship_type (PK)      │   │
└─────────────────────────────┘   │
  INDEX: (from_ticket)            │
  INDEX: (to_ticket)              │
  INDEX: (relationship_type)      │
                                  │
         (Many-to-Many) ◄─────────┘
```

**Diagram Legend:**
- **PK**: Primary Key
- **FK**: Foreign Key (logical, not enforced by constraints)
- **◄─┐**: Self-referential relationship
- **──►**: References another table
- **△**: One-to-many relationship
- **◄─►**: Many-to-many relationship

---

## Data Flow: Filesystem → Cache

```
┌──────────────────────┐
│  .pm/tickets/        │
│  ├── PROJ-1.md       │
│  ├── PROJ-2.md       │
│  └── PROJ-1/         │
│      └── comment.md  │
└──────────────────────┘
         │
         │ (1) ShouldSync() checks mtimes
         ▼
┌──────────────────────┐
│  cache_metadata      │
│  last_sync_timestamp │
└──────────────────────┘
         │
         │ (2) If stale, SyncCache()
         ▼
┌──────────────────────┐
│  Parse YAML + MD     │
│  - Extract metadata  │
│  - Parse comments    │
│  - Build relationships│
└──────────────────────┘
         │
         │ (3) Bulk insert in transaction
         ▼
┌──────────────────────┐
│  tickets             │
│  comments            │
│  relationships       │
│  cache_metadata      │  (update timestamp)
└──────────────────────┘
```

**Sync Process (from `internal/cache/sync.go`):**

1. **BEGIN TRANSACTION**
2. **DELETE** all rows from `tickets`, `comments`, `relationships`
3. **SCAN** `.pm/tickets/` directory:
   - Parse each `.md` file as a ticket
   - Parse comment directories (`TICKET-ID/*.md`)
   - Extract relationships from `depends_on` and `blocks` arrays
4. **BULK INSERT**:
   - Insert all tickets in a single query
   - Insert all comments in a single query
   - Insert all relationships in a single query
5. **UPDATE** `cache_metadata.last_sync_timestamp` to current time
6. **COMMIT TRANSACTION**

**Why Full Rebuild?**
- Simplifies logic (no need to track deltas)
- Handles edge cases (renamed files, deleted tickets)
- Fast enough for typical project sizes (&lt;10,000 tickets)
- Transactional safety prevents partial updates

---

## Synchronization Strategy

### Lazy Sync (Current Implementation)

**Trigger:** Before any read operation (`pm list`, `pm show`, etc.)

**Check:** `ShouldSync(pmPath)` compares:
```go
lastSyncTimestamp := cache_metadata['last_sync_timestamp']
for each ticket.md file {
    if file.ModTime >= lastSyncTimestamp {
        return true  // Needs sync
    }
}
return false  // Cache is fresh
```

**Timing Precision:**
- Filesystem mtimes truncated to **second precision**
- Comparison uses `!Before` (≥) to catch same-second modifications
- Prevents false negatives during rapid operations (tests, scripts)

**Performance:**
- **Cold start:** ~50-100ms to scan 100 ticket files
- **Warm cache:** &lt;5ms (no sync needed)
- **Full sync:** ~200-500ms for 100 tickets with comments

### Future: Manual Sync

Planned for Stage 3 (see AGENTS.md §7):
- `pm sync` - Force cache rebuild
- `--no-cache` flag - Bypass cache for debugging
- Cache health metrics and diagnostics

---

## Query Patterns

### Common Queries

**List all tickets by status:**
```sql
SELECT id, title, type, status, assignee, priority
FROM tickets
WHERE status = 'in-progress'
ORDER BY updated_at DESC;
```

**Get ticket with comments:**
```sql
-- Ticket metadata
SELECT * FROM tickets WHERE id = 'PROJ-123';

-- Comments for ticket
SELECT author, timestamp, filepath
FROM comments
WHERE ticket_id = 'PROJ-123'
ORDER BY timestamp ASC;
```

**Find blocked tickets:**
```sql
-- Tickets that depend on incomplete work
SELECT DISTINCT t.id, t.title, t.status
FROM tickets t
JOIN relationships r ON t.id = r.from_ticket
JOIN tickets dep ON r.to_ticket = dep.id
WHERE r.relationship_type = 'depends-on'
  AND dep.status != 'done';
```

**Get ticket hierarchy:**
```sql
-- Recursive CTE to build tree
WITH RECURSIVE ticket_tree AS (
  SELECT id, title, parent, 0 AS depth
  FROM tickets
  WHERE id = 'EPIC-1'
  
  UNION ALL
  
  SELECT t.id, t.title, t.parent, tt.depth + 1
  FROM tickets t
  JOIN ticket_tree tt ON t.parent = tt.id
)
SELECT * FROM ticket_tree ORDER BY depth, id;
```

---

## Validation and Constraints

### What's NOT Enforced by SQLite

To preserve flexibility and avoid filesystem/database sync issues, the cache uses **logical relationships** without foreign key constraints:

**No FK Constraints:**
- `tickets.parent` → `tickets.id` (allows forward references)
- `comments.ticket_id` → `tickets.id` (allows orphaned comments during transitions)
- `relationships.from_ticket` / `to_ticket` → `tickets.id`

**Why?**
- Tickets can reference not-yet-created dependencies
- Comments may briefly exist for deleted tickets (cleanup is eventual)
- Cache rebuild is always safe (full reconstruction from source)

### What IS Validated

**By Application Code (internal/ticket/validator.go):**
- Required fields (id, title, type, status, created_at)
- Enum validity (type ∈ {epic, story, task, bug})
- Status in workflow.yaml state list
- ID format matches `{PREFIX}-\d+`
- Date format is ISO8601 UTC
- No circular dependencies in `depends_on` graph

**By SQLite:**
- Primary key uniqueness
- NOT NULL constraints on required fields
- Index enforcement for query performance

---

## Migration Strategy

### Adding New Columns

When adding fields to tickets (e.g., `points`, `labels`):

1. **Create migration:**
   ```sql
   -- migrations/000005_add_points_field.up.sql
   ALTER TABLE tickets ADD COLUMN points INTEGER;
   ```

2. **Update sync logic:**
   ```go
   // internal/cache/sync.go
   tickets = append(tickets, ticketData{
       // ... existing fields ...
       points: t.Points,  // Add new field
   })
   ```

3. **Migration runs automatically:** Next command triggers lazy migration check

### Adding New Tables

Follow the pattern from `000003_add_comments_table.up.sql`:

1. CREATE TABLE with appropriate indexes
2. Update `SyncCache()` to populate new table
3. Add query methods in `internal/cache/`

### Rollback Strategy

```bash
# Down migrations exist for all schema changes
# Example: Roll back relationships table
migrate -path internal/migrations -database sqlite3://.pm/.cache.db down 1
```

**Note:** Down migrations are provided but rarely used (cache can be deleted and rebuilt from source).

---

## Testing

**Test Coverage (from `internal/cache/*_test.go`):**
- Migration up/down cycles
- Sync with empty directory
- Sync with existing tickets
- Staleness detection edge cases
- Timestamp precision handling
- Bulk insert performance
- Comment file parsing
- Relationship extraction

**Integration Tests:**
- Full workflow: init → create tickets → sync → query
- Concurrent access (multiple reads during sync)
- Cache corruption recovery (delete .cache.db → auto-rebuild)

---

## Performance Characteristics

### Benchmarks (Typical Hardware)

| Operation | 10 Tickets | 100 Tickets | 1000 Tickets |
|-----------|------------|-------------|--------------|
| ShouldSync check | &lt;1ms | ~5ms | ~50ms |
| Full sync | ~20ms | ~200ms | ~2s |
| List query | &lt;1ms | &lt;1ms | ~5ms |
| Show + comments | ~2ms | ~2ms | ~2ms |

**Bottlenecks:**
- **Filesystem I/O:** Scanning mtimes dominates ShouldSync
- **YAML parsing:** Dominates full sync
- **Bulk inserts:** Negligible (SQLite handles well)

**Optimizations:**
- Transaction batching for atomic updates
- Single bulk INSERT instead of row-by-row
- Indexes on frequently queried columns

---

## Maintenance

### Cache Corruption Recovery

**Symptoms:**
- `pm list` shows stale data
- Missing tickets that exist on filesystem
- SQLite errors

**Recovery:**
```bash
# Delete corrupted cache
rm .pm/.cache.db

# Next command auto-rebuilds
pm list
```

**Future:** `pm doctor` command to detect and fix issues

### Cache Location

**Path:** `.pm/.cache.db` (relative to project root)

**Git:** Ignored via `.pm/.gitignore`

**Sharing:** Each user has their own cache (never committed)

---

## Future Enhancements

### Full-Text Search (FTS)

**Planned for Stage 3:**
```sql
CREATE VIRTUAL TABLE tickets_fts USING fts5(
  id, title, body,
  content='tickets'
);
```

**Enables:**
- `pm search "authentication"` across title and body
- Relevance ranking
- Phrase matching

### Caching Comment Content

**Current:** Only comment metadata cached  
**Future:** Option to cache full content for offline search

**Trade-offs:**
- **Pros:** Faster `pm show`, enables comment search
- **Cons:** Larger cache size, sync complexity

### Analytics Tables

**Potential tables for reporting:**
- `ticket_state_history` - Track state transitions over time
- `assignee_stats` - Pre-aggregated workload metrics
- `sprint_snapshots` - Point-in-time ticket counts

---

## Summary

The SQLite cache is a **denormalized, read-only index** of the authoritative ticket markdown files. It trades storage space for query speed while maintaining eventual consistency through lazy synchronization. The schema is designed for:

- **Fast filtering** (tickets table with indexes)
- **Audit trails** (cache_metadata for sync tracking)
- **Collaboration** (comments table for discussion)
- **Dependency management** (relationships graph)

All tables are populated from the filesystem during sync operations, ensuring the cache is always a projection of the source files.
