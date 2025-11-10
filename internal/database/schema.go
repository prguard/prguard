package database

const schema = `
CREATE TABLE IF NOT EXISTS blocklist (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    reason TEXT NOT NULL,
    evidence_url TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    blocked_by TEXT NOT NULL,
    severity TEXT NOT NULL CHECK(severity IN ('low', 'medium', 'high')),
    source TEXT NOT NULL CHECK(source IN ('manual', 'imported', 'auto-detected')),
    metadata TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_blocklist_username ON blocklist(username);
CREATE INDEX IF NOT EXISTS idx_blocklist_severity ON blocklist(severity);
CREATE INDEX IF NOT EXISTS idx_blocklist_timestamp ON blocklist(timestamp);
`
