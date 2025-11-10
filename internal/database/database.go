package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prguard/prguard/pkg/models"
)

// DB wraps a database connection
type DB struct {
	conn *sql.DB
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(path string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates the database tables if they don't exist
func (db *DB) initSchema() error {
	_, err := db.conn.Exec(schema)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// AddEntry adds a new blocklist entry
func (db *DB) AddEntry(entry *models.BlocklistEntry) error {
	query := `
		INSERT INTO blocklist (id, username, reason, evidence_url, timestamp, blocked_by, severity, source, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		entry.ID,
		entry.Username,
		entry.Reason,
		entry.EvidenceURL,
		entry.Timestamp,
		entry.BlockedBy,
		entry.Severity,
		entry.Source,
		entry.Metadata,
	)
	return err
}

// GetEntry retrieves a blocklist entry by ID
func (db *DB) GetEntry(id string) (*models.BlocklistEntry, error) {
	query := `SELECT id, username, reason, evidence_url, timestamp, blocked_by, severity, source, metadata FROM blocklist WHERE id = ?`

	var entry models.BlocklistEntry
	err := db.conn.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.Username,
		&entry.Reason,
		&entry.EvidenceURL,
		&entry.Timestamp,
		&entry.BlockedBy,
		&entry.Severity,
		&entry.Source,
		&entry.Metadata,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// IsBlocked checks if a username is in the blocklist
func (db *DB) IsBlocked(username string) (bool, error) {
	query := `SELECT COUNT(*) FROM blocklist WHERE username = ?`
	var count int
	err := db.conn.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetEntriesByUsername retrieves all blocklist entries for a username
func (db *DB) GetEntriesByUsername(username string) ([]*models.BlocklistEntry, error) {
	query := `SELECT id, username, reason, evidence_url, timestamp, blocked_by, severity, source, metadata FROM blocklist WHERE username = ?`

	rows, err := db.conn.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.BlocklistEntry
	for rows.Next() {
		var entry models.BlocklistEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Username,
			&entry.Reason,
			&entry.EvidenceURL,
			&entry.Timestamp,
			&entry.BlockedBy,
			&entry.Severity,
			&entry.Source,
			&entry.Metadata,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	return entries, rows.Err()
}

// ListEntries retrieves all blocklist entries
func (db *DB) ListEntries() ([]*models.BlocklistEntry, error) {
	query := `SELECT id, username, reason, evidence_url, timestamp, blocked_by, severity, source, metadata FROM blocklist ORDER BY timestamp DESC`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.BlocklistEntry
	for rows.Next() {
		var entry models.BlocklistEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Username,
			&entry.Reason,
			&entry.EvidenceURL,
			&entry.Timestamp,
			&entry.BlockedBy,
			&entry.Severity,
			&entry.Source,
			&entry.Metadata,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	return entries, rows.Err()
}

// RemoveEntry removes a blocklist entry by ID
func (db *DB) RemoveEntry(id string) error {
	query := `DELETE FROM blocklist WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

// RemoveByUsername removes all blocklist entries for a username
func (db *DB) RemoveByUsername(username string) error {
	query := `DELETE FROM blocklist WHERE username = ?`
	_, err := db.conn.Exec(query, username)
	return err
}

// UpdateEntry updates an existing blocklist entry
func (db *DB) UpdateEntry(entry *models.BlocklistEntry) error {
	query := `
		UPDATE blocklist
		SET reason = ?, evidence_url = ?, severity = ?, metadata = ?
		WHERE id = ?
	`
	_, err := db.conn.Exec(query, entry.Reason, entry.EvidenceURL, entry.Severity, entry.Metadata, entry.ID)
	return err
}
