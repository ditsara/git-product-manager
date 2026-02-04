CREATE TABLE comments (
  ticket_id TEXT NOT NULL,
  author TEXT NOT NULL,
  timestamp TEXT NOT NULL,
  filepath TEXT NOT NULL,
  PRIMARY KEY (ticket_id, timestamp, author)
);

CREATE INDEX idx_ticket_comments ON comments(ticket_id, timestamp);
