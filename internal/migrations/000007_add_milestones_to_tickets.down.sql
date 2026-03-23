-- SQLite does not support DROP COLUMN (before version 3.35.0).
-- Manual rollback: recreate tickets table without the milestones column.
SELECT 1;
