// Copyright 2025 Logan Lindquist Land
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blocklist

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

// Manager handles blocklist operations
type Manager struct {
	db *database.DB
}

// NewManager creates a new blocklist manager
func NewManager(db *database.DB) *Manager {
	return &Manager{db: db}
}

// Block adds a user to the blocklist
func (m *Manager) Block(username, reason, evidenceURL, blockedBy, severity, source string) (*models.BlocklistEntry, error) {
	entry := models.NewBlocklistEntry(username, reason, evidenceURL, blockedBy, severity, source)
	if err := m.db.AddEntry(entry); err != nil {
		return nil, fmt.Errorf("failed to add blocklist entry: %w", err)
	}
	return entry, nil
}

// Unblock removes a user from the blocklist
func (m *Manager) Unblock(username string) error {
	return m.db.RemoveByUsername(username)
}

// IsBlocked checks if a user is blocked
func (m *Manager) IsBlocked(username string) (bool, error) {
	return m.db.IsBlocked(username)
}

// List returns all blocklist entries
func (m *Manager) List() ([]*models.BlocklistEntry, error) {
	return m.db.ListEntries()
}

// GetByUsername returns all blocklist entries for a specific user
func (m *Manager) GetByUsername(username string) ([]*models.BlocklistEntry, error) {
	return m.db.GetEntriesByUsername(username)
}

// ExportJSON exports the blocklist to a JSON file
func (m *Manager) ExportJSON(path string) error {
	entries, err := m.db.ListEntries()
	if err != nil {
		return fmt.Errorf("failed to get entries: %w", err)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportCSV exports the blocklist to a CSV file
func (m *Manager) ExportCSV(path string) error {
	entries, err := m.db.ListEntries()
	if err != nil {
		return fmt.Errorf("failed to get entries: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"ID", "Username", "Reason", "EvidenceURL", "Timestamp", "BlockedBy", "Severity", "Source"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		record := []string{
			entry.ID,
			entry.Username,
			entry.Reason,
			entry.EvidenceURL,
			entry.Timestamp.Format(time.RFC3339),
			entry.BlockedBy,
			entry.Severity,
			entry.Source,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// ImportJSON imports blocklist entries from a JSON file
func (m *Manager) ImportJSON(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	var entries []*models.BlocklistEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return m.importEntries(entries)
}

// ImportJSONFromURL imports blocklist entries from a remote JSON URL
func (m *Manager) ImportJSONFromURL(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var entries []*models.BlocklistEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return m.importEntries(entries)
}

// importEntries imports a slice of entries with deduplication
func (m *Manager) importEntries(entries []*models.BlocklistEntry) (int, error) {
	imported := 0
	for _, entry := range entries {
		// Check if entry already exists by ID
		existing, err := m.db.GetEntry(entry.ID)
		if err != nil {
			return imported, fmt.Errorf("failed to check for existing entry: %w", err)
		}

		if existing != nil {
			// Entry exists, skip or update based on severity
			if shouldUpdate(existing, entry) {
				if err := m.db.UpdateEntry(entry); err != nil {
					return imported, fmt.Errorf("failed to update entry: %w", err)
				}
				imported++
			}
			continue
		}

		// Add new entry
		entry.Source = models.SourceImported
		if err := m.db.AddEntry(entry); err != nil {
			return imported, fmt.Errorf("failed to add entry: %w", err)
		}
		imported++
	}

	return imported, nil
}

// shouldUpdate determines if an existing entry should be updated with new data
func shouldUpdate(existing, new *models.BlocklistEntry) bool {
	// Update if new entry has higher severity
	severityMap := map[string]int{
		models.SeverityLow:    1,
		models.SeverityMedium: 2,
		models.SeverityHigh:   3,
	}
	return severityMap[new.Severity] > severityMap[existing.Severity]
}
