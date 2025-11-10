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

package database

import (
	"testing"
	"time"

	"github.com/prguard/prguard/pkg/models"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *DB {
	db, err := NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}

func TestAddEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	entry := models.NewBlocklistEntry(
		"spammer123",
		"Multiple spam PRs",
		"https://github.com/org/repo/pull/123",
		"maintainer",
		models.SeverityHigh,
		models.SourceManual,
	)

	err := db.AddEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Verify entry was added
	retrieved, err := db.GetEntry(entry.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Entry not found")
	}

	if retrieved.Username != "spammer123" {
		t.Errorf("Expected username 'spammer123', got '%s'", retrieved.Username)
	}
}

func TestIsBlocked(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// User should not be blocked initially
	blocked, err := db.IsBlocked("testuser")
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if blocked {
		t.Error("User should not be blocked initially")
	}

	// Add entry
	entry := models.NewBlocklistEntry(
		"testuser",
		"Test reason",
		"https://example.com",
		"admin",
		models.SeverityMedium,
		models.SourceManual,
	)
	db.AddEntry(entry)

	// User should now be blocked
	blocked, err = db.IsBlocked("testuser")
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if !blocked {
		t.Error("User should be blocked after adding entry")
	}
}

func TestGetEntriesByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	username := "multientry"

	// Add multiple entries for same user
	entry1 := models.NewBlocklistEntry(username, "Reason 1", "https://example.com/1", "admin", models.SeverityLow, models.SourceManual)
	entry2 := models.NewBlocklistEntry(username, "Reason 2", "https://example.com/2", "admin", models.SeverityHigh, models.SourceImported)

	db.AddEntry(entry1)
	db.AddEntry(entry2)

	entries, err := db.GetEntriesByUsername(username)
	if err != nil {
		t.Fatalf("GetEntriesByUsername failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestListEntries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Add test entries
	for i := 0; i < 3; i++ {
		entry := models.NewBlocklistEntry(
			"user"+string(rune(i+'0')),
			"Test reason",
			"https://example.com",
			"admin",
			models.SeverityMedium,
			models.SourceManual,
		)
		db.AddEntry(entry)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	entries, err := db.ListEntries()
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Verify entries are sorted by timestamp descending
	for i := 0; i < len(entries)-1; i++ {
		if entries[i].Timestamp.Before(entries[i+1].Timestamp) {
			t.Error("Entries are not sorted by timestamp descending")
		}
	}
}

func TestRemoveEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	entry := models.NewBlocklistEntry("toremove", "Test", "https://example.com", "admin", models.SeverityLow, models.SourceManual)
	db.AddEntry(entry)

	// Verify entry exists
	blocked, _ := db.IsBlocked("toremove")
	if !blocked {
		t.Fatal("Entry was not added")
	}

	// Remove by ID
	err := db.RemoveEntry(entry.ID)
	if err != nil {
		t.Fatalf("RemoveEntry failed: %v", err)
	}

	// Verify entry is gone
	blocked, _ = db.IsBlocked("toremove")
	if blocked {
		t.Error("Entry was not removed")
	}
}

func TestRemoveByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	username := "multiremove"

	// Add multiple entries
	entry1 := models.NewBlocklistEntry(username, "Reason 1", "https://example.com/1", "admin", models.SeverityLow, models.SourceManual)
	entry2 := models.NewBlocklistEntry(username, "Reason 2", "https://example.com/2", "admin", models.SeverityHigh, models.SourceImported)

	db.AddEntry(entry1)
	db.AddEntry(entry2)

	// Remove all by username
	err := db.RemoveByUsername(username)
	if err != nil {
		t.Fatalf("RemoveByUsername failed: %v", err)
	}

	// Verify all entries are gone
	entries, _ := db.GetEntriesByUsername(username)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after removal, got %d", len(entries))
	}
}

func TestUpdateEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	entry := models.NewBlocklistEntry("updatetest", "Original reason", "https://example.com", "admin", models.SeverityLow, models.SourceManual)
	db.AddEntry(entry)

	// Update the entry
	entry.Reason = "Updated reason"
	entry.Severity = models.SeverityHigh

	err := db.UpdateEntry(entry)
	if err != nil {
		t.Fatalf("UpdateEntry failed: %v", err)
	}

	// Retrieve and verify
	updated, err := db.GetEntry(entry.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated entry: %v", err)
	}

	if updated.Reason != "Updated reason" {
		t.Errorf("Expected reason 'Updated reason', got '%s'", updated.Reason)
	}

	if updated.Severity != models.SeverityHigh {
		t.Errorf("Expected severity 'high', got '%s'", updated.Severity)
	}
}

func TestSeverityConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	entry := &models.BlocklistEntry{
		ID:          "test-id",
		Username:    "test",
		Reason:      "test",
		EvidenceURL: "https://example.com",
		Timestamp:   time.Now(),
		BlockedBy:   "admin",
		Severity:    "invalid-severity", // Invalid severity
		Source:      models.SourceManual,
		Metadata:    "{}",
	}

	err := db.AddEntry(entry)
	if err == nil {
		t.Error("Expected error for invalid severity, got nil")
	}
}

func TestSourceConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	entry := &models.BlocklistEntry{
		ID:          "test-id",
		Username:    "test",
		Reason:      "test",
		EvidenceURL: "https://example.com",
		Timestamp:   time.Now(),
		BlockedBy:   "admin",
		Severity:    models.SeverityMedium,
		Source:      "invalid-source", // Invalid source
		Metadata:    "{}",
	}

	err := db.AddEntry(entry)
	if err == nil {
		t.Error("Expected error for invalid source, got nil")
	}
}
