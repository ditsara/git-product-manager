CREATE TABLE milestones (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT,
  due_date TEXT,
  state TEXT NOT NULL,
  created_at TEXT NOT NULL,
  closed_at TEXT
);

CREATE INDEX idx_milestone_state ON milestones(state);
