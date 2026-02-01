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
