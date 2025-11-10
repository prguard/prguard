package commands

import (
	"path/filepath"
	"testing"

	"github.com/prguard/prguard/internal/blocklist"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

func TestListCommand_Empty(t *testing.T) {
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

	// List should succeed with empty database
	err = runList(configPath)
	if err != nil {
		t.Errorf("runList failed: %v", err)
	}
}

func TestListCommand_WithEntries(t *testing.T) {
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

	// Add some test entries
	manager := blocklist.NewManager(db)
	testUsers := []struct {
		username string
		reason   string
		severity string
	}{
		{"spammer1", "spam PRs", models.SeverityHigh},
		{"spammer2", "fake contributions", models.SeverityMedium},
		{"spammer3", "readme spam", models.SeverityLow},
	}

	for _, u := range testUsers {
		_, err := manager.Block(u.username, u.reason, "https://github.com/test/repo/pull/1", "testowner", u.severity, models.SourceManual)
		if err != nil {
			t.Fatalf("failed to add test user %s: %v", u.username, err)
		}
	}

	// List should succeed and show all entries
	err = runList(configPath)
	if err != nil {
		t.Errorf("runList failed: %v", err)
	}

	// Verify entries exist in database
	entries, err := manager.List()
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(entries) != len(testUsers) {
		t.Errorf("expected %d entries, got %d", len(testUsers), len(entries))
	}

	// Verify each user is in the list
	usernames := make(map[string]bool)
	for _, entry := range entries {
		usernames[entry.Username] = true
	}

	for _, u := range testUsers {
		if !usernames[u.username] {
			t.Errorf("expected user %s in blocklist", u.username)
		}
	}
}

func TestListCommand_MissingConfig(t *testing.T) {
	configPath := "/nonexistent/config.yaml"
	err := runList(configPath)
	if err == nil {
		t.Error("expected error with missing config")
	}
}
