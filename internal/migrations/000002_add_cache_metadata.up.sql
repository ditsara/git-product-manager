CREATE TABLE cache_metadata (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

INSERT INTO cache_metadata (key, value) VALUES ('last_sync_timestamp', '1970-01-01T00:00:00Z');
