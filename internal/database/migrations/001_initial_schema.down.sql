-- Rollback initial schema
DROP INDEX IF EXISTS idx_blocklist_timestamp;
DROP INDEX IF EXISTS idx_blocklist_severity;
DROP INDEX IF EXISTS idx_blocklist_username;
DROP TABLE IF EXISTS blocklist;
