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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

func setupTestManager(t *testing.T) (*Manager, *database.DB) {
	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return NewManager(db), db
}

func TestBlock(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	entry, err := manager.Block(
		"spammer",
		"Multiple spam PRs",
		"https://github.com/org/repo/pull/123",
		"maintainer",
		models.SeverityHigh,
		models.SourceManual,
	)

	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}

	if entry.Username != "spammer" {
		t.Errorf("Expected username 'spammer', got '%s'", entry.Username)
	}

	if entry.Severity != models.SeverityHigh {
		t.Errorf("Expected severity 'high', got '%s'", entry.Severity)
	}

	// Verify user is blocked
	blocked, err := manager.IsBlocked("spammer")
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if !blocked {
		t.Error("User should be blocked")
	}
}

func TestUnblock(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	// Block a user
	manager.Block("toremove", "test", "https://example.com", "admin", models.SeverityLow, models.SourceManual)

	// Verify blocked
	blocked, _ := manager.IsBlocked("toremove")
	if !blocked {
		t.Fatal("User was not blocked")
	}

	// Unblock
	err := manager.Unblock("toremove")
	if err != nil {
		t.Fatalf("Unblock failed: %v", err)
	}

	// Verify unblocked
	blocked, _ = manager.IsBlocked("toremove")
	if blocked {
		t.Error("User should be unblocked")
	}
}

func TestExportJSON(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	// Create temp directory
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "blocklist.json")

	// Add test entries
	manager.Block("user1", "reason1", "https://example.com/1", "admin", models.SeverityHigh, models.SourceManual)
	manager.Block("user2", "reason2", "https://example.com/2", "admin", models.SeverityMedium, models.SourceImported)

	// Export
	err := manager.ExportJSON(exportPath)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatal("Export file was not created")
	}

	// Read and parse JSON
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	var entries []*models.BlocklistEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in export, got %d", len(entries))
	}
}

func TestExportCSV(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "blocklist.csv")

	// Add test entries
	manager.Block("csvuser1", "test reason", "https://example.com", "admin", models.SeverityLow, models.SourceManual)

	// Export
	err := manager.ExportCSV(exportPath)
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	content := string(data)
	if content == "" {
		t.Error("CSV file is empty")
	}

	// Check for header
	if !contains(content, "Username") {
		t.Error("CSV file missing header")
	}

	// Check for data
	if !contains(content, "csvuser1") {
		t.Error("CSV file missing user data")
	}
}

func TestImportJSON(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	tmpDir := t.TempDir()
	importPath := filepath.Join(tmpDir, "import.json")

	// Create test data
	testEntries := []*models.BlocklistEntry{
		models.NewBlocklistEntry("import1", "reason1", "https://example.com/1", "admin", models.SeverityHigh, models.SourceImported),
		models.NewBlocklistEntry("import2", "reason2", "https://example.com/2", "admin", models.SeverityMedium, models.SourceImported),
	}

	// Write JSON file
	data, _ := json.Marshal(testEntries)
	os.WriteFile(importPath, data, 0644)

	// Import
	count, err := manager.ImportJSON(importPath)
	if err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 entries imported, got %d", count)
	}

	// Verify entries were imported
	entries, _ := manager.List()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in database, got %d", len(entries))
	}

	// Verify specific users
	blocked, _ := manager.IsBlocked("import1")
	if !blocked {
		t.Error("import1 should be blocked")
	}

	blocked, _ = manager.IsBlocked("import2")
	if !blocked {
		t.Error("import2 should be blocked")
	}
}

func TestImportJSON_Deduplication(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	tmpDir := t.TempDir()

	// Create an entry with a specific ID
	originalEntry := models.NewBlocklistEntry("dupuser", "original reason", "https://example.com", "admin", models.SeverityLow, models.SourceManual)
	db.AddEntry(originalEntry)

	// Create import file with same ID but higher severity
	updatedEntry := &models.BlocklistEntry{
		ID:          originalEntry.ID, // Same ID
		Username:    "dupuser",
		Reason:      "updated reason",
		EvidenceURL: "https://example.com/updated",
		Timestamp:   originalEntry.Timestamp,
		BlockedBy:   "admin",
		Severity:    models.SeverityHigh, // Higher severity
		Source:      models.SourceImported,
		Metadata:    "{}",
	}

	importPath := filepath.Join(tmpDir, "dup.json")
	data, _ := json.Marshal([]*models.BlocklistEntry{updatedEntry})
	os.WriteFile(importPath, data, 0644)

	// Import - should update existing entry
	count, err := manager.ImportJSON(importPath)
	if err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 entry updated, got %d", count)
	}

	// Verify only one entry exists with updated severity
	entries, _ := manager.GetByUsername("dupuser")
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for dupuser, got %d", len(entries))
	}

	if entries[0].Severity != models.SeverityHigh {
		t.Errorf("Expected severity to be updated to 'high', got '%s'", entries[0].Severity)
	}
}

func TestImportJSON_LowerSeverityIgnored(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	tmpDir := t.TempDir()

	// Create high severity entry
	originalEntry := models.NewBlocklistEntry("severityuser", "original", "https://example.com", "admin", models.SeverityHigh, models.SourceManual)
	db.AddEntry(originalEntry)

	// Try to import with lower severity
	lowerSeverityEntry := &models.BlocklistEntry{
		ID:          originalEntry.ID,
		Username:    "severityuser",
		Reason:      "updated",
		EvidenceURL: "https://example.com",
		Timestamp:   originalEntry.Timestamp,
		BlockedBy:   "admin",
		Severity:    models.SeverityLow, // Lower severity
		Source:      models.SourceImported,
		Metadata:    "{}",
	}

	importPath := filepath.Join(tmpDir, "lower.json")
	data, _ := json.Marshal([]*models.BlocklistEntry{lowerSeverityEntry})
	os.WriteFile(importPath, data, 0644)

	// Import - should not update
	count, _ := manager.ImportJSON(importPath)

	// Entry should not be counted as imported since severity is lower
	if count != 0 {
		t.Errorf("Expected 0 entries updated (lower severity), got %d", count)
	}

	// Verify severity remains high
	entry, _ := db.GetEntry(originalEntry.ID)
	if entry.Severity != models.SeverityHigh {
		t.Errorf("Expected severity to remain 'high', got '%s'", entry.Severity)
	}
}

func TestGetByUsername(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	username := "multientry"

	// Add multiple entries for same user
	manager.Block(username, "reason1", "https://example.com/1", "admin", models.SeverityLow, models.SourceManual)
	manager.Block(username, "reason2", "https://example.com/2", "admin", models.SeverityHigh, models.SourceImported)

	entries, err := manager.GetByUsername(username)
	if err != nil {
		t.Fatalf("GetByUsername failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestList(t *testing.T) {
	manager, db := setupTestManager(t)
	defer db.Close()

	// Add multiple users
	manager.Block("user1", "reason1", "https://example.com/1", "admin", models.SeverityLow, models.SourceManual)
	manager.Block("user2", "reason2", "https://example.com/2", "admin", models.SeverityMedium, models.SourceManual)
	manager.Block("user3", "reason3", "https://example.com/3", "admin", models.SeverityHigh, models.SourceManual)

	entries, err := manager.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
