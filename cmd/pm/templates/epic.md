---
id: {{.ID}}
title: "{{.Title}}"
type: epic
status: backlog  # Current workflow state
priority: medium  # low, medium, high, critical
points: 0  # Story points for estimation

# Relationships - use ticket IDs (e.g., PROJ-123)
parent: "{{.Parent}}"  # Parent epic (for nested epics)
depends_on: []  # Must complete these first
blocks: []  # This blocks these tickets
related: []  # Related work (duplicates, see-also)

labels: []  # Tags from labels.yaml
milestones: []  # Milestone IDs this ticket belongs to
assignee: ""  # GitHub username or email
created_at: "{{.CreatedAt}}"
updated_at: "{{.UpdatedAt}}"
---

# Description

