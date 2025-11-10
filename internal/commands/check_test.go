package commands

import (
	"path/filepath"
	"testing"

	"github.com/prguard/prguard/internal/blocklist"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

func TestCheckCommand_UserBlocked(t *testing.T) {
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

	// Add a blocked user
	manager := blocklist.NewManager(db)
	_, err = manager.Block("blockeduser", "spam", "https://github.com/test/repo/pull/1", "test-org", models.SeverityHigh, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to block user: %v", err)
	}

	// Check should report user as blocked
	err = runCheck(configPath, "blockeduser")
	if err != nil {
		t.Errorf("runCheck failed: %v", err)
	}
}

func TestCheckCommand_UserNotBlocked(t *testing.T) {
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

	// Check should report user as not blocked
	err = runCheck(configPath, "normaluser")
	if err != nil {
		t.Errorf("runCheck failed: %v", err)
	}
}

func TestCheckCommand_MultipleEntries(t *testing.T) {
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

	// Add multiple entries for the same user
	manager := blocklist.NewManager(db)
	_, err = manager.Block("repeatoffender", "spam PR 1", "https://github.com/test/repo/pull/1", "test-org", models.SeverityLow, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to block user first time: %v", err)
	}

	_, err = manager.Block("repeatoffender", "spam PR 2", "https://github.com/test/repo/pull/2", "test-org", models.SeverityHigh, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to block user second time: %v", err)
	}

	// Check should show both entries
	err = runCheck(configPath, "repeatoffender")
	if err != nil {
		t.Errorf("runCheck failed: %v", err)
	}

	// Verify both entries exist
	entries, err := manager.GetByUsername("repeatoffender")
	if err != nil {
		t.Fatalf("failed to get entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}
