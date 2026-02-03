---
id: GPM-14
title: "Implement Milestones for grouping and tracking tickets"
type: epic
status: backlog
priority: low
points: 0

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: ""
depends_on: [GPM-2]
blocks: []
related: []

labels: [feature, milestones, planning]
assignee: ""
created_at: "2026-02-03T04:13:05Z"
updated_at: "2026-02-03T04:13:05Z"
---

# Description

[Sonnet 4.5]

**Epic:** Add support for Milestones to arbitrarily group tickets and track progress toward a target date.

**Note:** This feature should be implemented after Stage 3 (GPM-2) is complete, as it builds on the relationship and visualization capabilities.

## Overview

Milestones provide a way to group related work toward a specific goal or deadline. Unlike epics (which represent a body of work), milestones represent a point in time with associated deliverables.

### GitLab Milestones Inspiration

In GitLab, milestones allow you to:
- Group issues and merge requests under a common goal
- Set a due date for the milestone
- Track progress (% complete based on closed vs total items)
- Filter and view all work associated with a milestone
- Close milestones when the goal is reached

## Use Cases

1. **Release Planning**: "v1.0 Release" milestone with target date Feb 28, 2026
2. **Sprint Planning**: "Sprint 3" milestone with 2-week duration
3. **Project Phases**: "MVP Launch" milestone grouping all required features
4. **Time-boxed Goals**: "Q1 Security Improvements" with quarterly deadline

## Proposed Design

### Data Model

Milestones stored as YAML files in `.pm/milestones/`:

```yaml
# .pm/milestones/v1-0-release.yaml
id: v1-0-release
title: "Version 1.0 Release"
description: "First stable release with core features"
due_date: "2026-02-28"
state: active  # active, closed
created_at: "2026-01-15T10:00:00Z"
closed_at: null  # Set when milestone is closed
```

### Ticket Integration

Add `milestone` field to ticket YAML:

```yaml
---
id: GPM-1
title: "Stage 2: Collaboration and History"
milestone: v1-0-release  # Links to milestone ID
# ... other fields
---
```

### Commands

```bash
# Create milestone
pm milestone create "v1.0 Release" --due 2026-02-28

# List milestones
pm milestone list
# Output:
# v1-0-release    Version 1.0 Release    Feb 28, 2026    5/12 closed (42%)

# Show milestone details
pm milestone show v1-0-release
# Shows: description, due date, list of tickets, progress

# Assign ticket to milestone
pm edit GPM-1 --field milestone=v1-0-release

# List tickets in milestone
pm list --milestone v1-0-release

# Close milestone
pm milestone close v1-0-release

# Burndown view (optional advanced feature)
pm milestone burndown v1-0-release
```

### Progress Tracking

Calculate progress automatically:
- Total tickets assigned to milestone
- Tickets in "done" state
- Percentage complete
- Days remaining until due date
- Burndown chart (optional visualization)

## File Structure

```
.pm/
├── milestones/
│   ├── v1-0-release.yaml
│   ├── sprint-3.yaml
│   └── mvp-launch.yaml
├── tickets/
│   ├── GPM-1.md  (contains: milestone: v1-0-release)
│   └── GPM-2.md  (contains: milestone: mvp-launch)
```

## Implementation Phases

### Phase 1: Basic Milestones
- [ ] Create `.pm/milestones/` directory structure
- [ ] Add `milestone` field to ticket schema
- [ ] Implement `pm milestone create/list/show/close`
- [ ] Add `pm list --milestone <id>` filtering
- [ ] Validation: milestone ID must exist when referenced

### Phase 2: Progress Tracking
- [ ] Calculate completion percentage
- [ ] Show days remaining vs due date
- [ ] `pm milestone show` displays progress bar
- [ ] Warning when milestone is overdue
- [ ] List overdue milestones

### Phase 3: Advanced Features
- [ ] Milestone templates (sprint, release, etc.)
- [ ] Burndown chart visualization (ASCII art or web export)
- [ ] Milestone dependencies (one milestone blocks another)
- [ ] Historical view: closed milestones archive
- [ ] Export milestone report to Markdown

## Database Schema (Cache)

Add milestones support to SQLite cache:

```sql
CREATE TABLE milestones (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT,
  due_date TEXT,
  state TEXT NOT NULL,
  created_at TEXT NOT NULL,
  closed_at TEXT
);

-- Update tickets table
ALTER TABLE tickets ADD COLUMN milestone TEXT;

-- Index for fast filtering
CREATE INDEX idx_ticket_milestone ON tickets(milestone);
```

## Acceptance Criteria

- [ ] Milestones can be created with title and due date
- [ ] Tickets can be assigned to milestones
- [ ] `pm list --milestone <id>` shows only tickets in that milestone
- [ ] Progress percentage calculated correctly
- [ ] Milestones are version controlled (YAML files in git)
- [ ] Cache syncs milestone data for fast queries
- [ ] Overdue milestones are visually indicated
- [ ] Can close milestones when complete

## Why After Stage 3?

This feature depends on:
- **Filtering capabilities** (from Stage 3 search)
- **Relationship tracking** (milestone ↔ ticket associations)
- **Visualization** (progress bars, potentially charts)
- **Robust caching** (efficient milestone queries)

Once Stage 3 is complete, the foundation for milestones will be solid.

## Future Enhancements

- Integration with git tags (auto-milestone for releases)
- Email/notification when milestone due date approaches
- Milestone templates (e.g., "2-week sprint" auto-generates dates)
- Web UI for milestone gantt charts
- Export to project management formats (CSV, JSON)

