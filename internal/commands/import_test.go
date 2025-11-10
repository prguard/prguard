package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/prguard/prguard/internal/blocklist"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

func TestImportCommand_FromFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")
	importPath := filepath.Join(tempDir, "import.json")

	// Create test config
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Token: "test-token",
			Org:   "test-org",
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: dbPath,
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Initialize database
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create import file with test data
	testEntries := []models.BlocklistEntry{
		{
			ID:          "test-id-1",
			Username:    "importuser1",
			Reason:      "spam",
			EvidenceURL: "https://github.com/test/repo/pull/1",
			BlockedBy:   "test-org",
			Severity:    models.SeverityMedium,
			Source:      models.SourceImported,
		},
		{
			ID:          "test-id-2",
			Username:    "importuser2",
			Reason:      "abuse",
			EvidenceURL: "https://github.com/test/repo/pull/2",
			BlockedBy:   "test-org",
			Severity:    models.SeverityHigh,
			Source:      models.SourceImported,
		},
	}

	data, err := json.Marshal(testEntries)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(importPath, data, 0644); err != nil {
		t.Fatalf("failed to write import file: %v", err)
	}

	// Import from file
	err = runImport(configPath, importPath, "")
	if err != nil {
		t.Errorf("runImport failed: %v", err)
	}

	// Verify entries were imported
	manager := blocklist.NewManager(db)
	entries, err := manager.List()
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(entries) != len(testEntries) {
		t.Errorf("expected %d entries after import, got %d", len(testEntries), len(entries))
	}

	// Verify usernames
	usernames := make(map[string]bool)
	for _, entry := range entries {
		usernames[entry.Username] = true
	}

	for _, testEntry := range testEntries {
		if !usernames[testEntry.Username] {
			t.Errorf("expected user %s in imported entries", testEntry.Username)
		}
	}
}

func TestImportCommand_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")
	importPath := filepath.Join(tempDir, "import.json")

	// Create test config
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Token: "test-token",
			User:  "testowner",
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: dbPath,
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Initialize database
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add an existing entry
	manager := blocklist.NewManager(db)
	_, err = manager.Block("existinguser", "spam", "https://github.com/test/repo/pull/1", "testowner", models.SeverityLow, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to add existing entry: %v", err)
	}

	// Create import file with duplicate entry (same ID, higher severity)
	existingEntry, err := manager.GetByUsername("existinguser")
	if err != nil {
		t.Fatalf("failed to get existing entry: %v", err)
	}
	if len(existingEntry) == 0 {
		t.Fatal("existing entry not found")
	}

	duplicateEntry := existingEntry[0]
	duplicateEntry.Severity = models.SeverityHigh // Higher severity should be kept
	duplicateEntry.Reason = "updated reason"

	testEntries := []models.BlocklistEntry{*duplicateEntry}

	data, err := json.Marshal(testEntries)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(importPath, data, 0644); err != nil {
		t.Fatalf("failed to write import file: %v", err)
	}

	// Import (should deduplicate)
	err = runImport(configPath, importPath, "")
	if err != nil {
		t.Errorf("runImport failed: %v", err)
	}

	// Verify only one entry exists with higher severity
	entries, err := manager.GetByUsername("existinguser")
	if err != nil {
		t.Fatalf("failed to get entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry after deduplication, got %d", len(entries))
	}

	if entries[0].Severity != models.SeverityHigh {
		t.Errorf("expected severity to be updated to high, got %s", entries[0].Severity)
	}
}

func TestImportCommand_MissingFlags(t *testing.T) {
	configPath := "config.yaml"

	// No file or URL specified
	err := runImport(configPath, "", "")
	if err == nil {
		t.Error("expected error when neither file nor URL specified")
	}
}

func TestImportCommand_BothFlags(t *testing.T) {
	configPath := "config.yaml"

	// Both file and URL specified
	err := runImport(configPath, "file.json", "http://example.com/blocklist.json")
	if err == nil {
		t.Error("expected error when both file and URL specified")
	}
}

func TestImportCommand_NonexistentFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create test config
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Token: "test-token",
			Org:   "test-org",
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: dbPath,
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Initialize database
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Try to import from nonexistent file
	err = runImport(configPath, "/nonexistent/file.json", "")
	if err == nil {
		t.Error("expected error with nonexistent file")
	}
}

func TestImportCommand_InvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")
	importPath := filepath.Join(tempDir, "invalid.json")

	// Create test config
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Token: "test-token",
			Org:   "test-org",
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: dbPath,
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Initialize database
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create invalid JSON file
	if err := os.WriteFile(importPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid JSON: %v", err)
	}

	// Try to import invalid JSON
	err = runImport(configPath, importPath, "")
	if err == nil {
		t.Error("expected error with invalid JSON")
	}
}

func TestImportCommand_EmptyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")
	importPath := filepath.Join(tempDir, "empty.json")

	// Create test config
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Token: "test-token",
			Org:   "test-org",
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: dbPath,
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Initialize database
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create empty JSON array
	if err := os.WriteFile(importPath, []byte("[]"), 0644); err != nil {
		t.Fatalf("failed to write empty JSON: %v", err)
	}

	// Import empty file
	err = runImport(configPath, importPath, "")
	if err != nil {
		t.Errorf("runImport with empty file failed: %v", err)
	}

	// Verify no entries were imported
	manager := blocklist.NewManager(db)
	entries, err := manager.List()
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries after importing empty file, got %d", len(entries))
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		singular string
		plural   string
		count    int
		expected string
	}{
		{"entry", "entries", 0, "entries"},
		{"entry", "entries", 1, "entry"},
		{"entry", "entries", 2, "entries"},
		{"entry", "entries", 100, "entries"},
		{"user", "users", 1, "user"},
		{"user", "users", 5, "users"},
	}

	for _, tt := range tests {
		result := pluralize(tt.singular, tt.plural, tt.count)
		if result != tt.expected {
			t.Errorf("pluralize(%s, %s, %d) = %s, expected %s",
				tt.singular, tt.plural, tt.count, result, tt.expected)
		}
	}
}
