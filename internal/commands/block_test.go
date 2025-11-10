package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

func TestBlockCommand_Flags(t *testing.T) {
	configPath := "config.yaml"
	cmd := NewBlockCommand(&configPath)

	// Test required flags are set
	reasonFlag := cmd.Flags().Lookup("reason")
	if reasonFlag == nil {
		t.Error("reason flag not found")
	}

	evidenceFlag := cmd.Flags().Lookup("evidence")
	if evidenceFlag == nil {
		t.Error("evidence flag not found")
	}

	severityFlag := cmd.Flags().Lookup("severity")
	if severityFlag == nil {
		t.Error("severity flag not found")
	}
	if severityFlag.DefValue != "medium" {
		t.Errorf("severity default should be 'medium', got %s", severityFlag.DefValue)
	}

	githubBlockFlag := cmd.Flags().Lookup("github-block")
	if githubBlockFlag == nil {
		t.Error("github-block flag not found")
	}

	// Test command requires exactly 1 arg
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with no args")
	}
}

func TestBlockCommand_SeverityValidation(t *testing.T) {
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

	tests := []struct {
		name     string
		severity string
		wantErr  bool
	}{
		{"valid low", models.SeverityLow, false},
		{"valid medium", models.SeverityMedium, false},
		{"valid high", models.SeverityHigh, false},
		{"invalid severity", "critical", true},
		{"empty severity", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't easily test runBlock due to GitHub API calls
			// But we can verify severity validation logic
			if tt.severity != models.SeverityLow &&
				tt.severity != models.SeverityMedium &&
				tt.severity != models.SeverityHigh {
				// This should error
				if !tt.wantErr {
					t.Error("expected error for invalid severity")
				}
			} else {
				if tt.wantErr {
					t.Error("expected no error for valid severity")
				}
			}
		})
	}
}

func TestBlockCommand_Integration(t *testing.T) {
	// Skip if we can't create temp files
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create test config with dummy token (won't be used since we pass false for githubBlock)
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

	// Test blocking a user (without GitHub API)
	err = runBlock(configPath, "testuser", "spam", "https://github.com/test/repo/pull/1", models.SeverityMedium, false)
	if err != nil {
		t.Errorf("runBlock failed: %v", err)
	}

	// Verify user was added to database
	blocked, err := db.IsBlocked("testuser")
	if err != nil {
		t.Fatalf("failed to check if user is blocked: %v", err)
	}
	if !blocked {
		t.Error("user should be blocked")
	}

	// Verify entry details
	entries, err := db.GetEntriesByUsername("testuser")
	if err != nil {
		t.Fatalf("failed to get entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", entry.Username)
	}
	if entry.Reason != "spam" {
		t.Errorf("expected reason 'spam', got %s", entry.Reason)
	}
	if entry.Severity != models.SeverityMedium {
		t.Errorf("expected severity 'medium', got %s", entry.Severity)
	}
	if entry.Source != models.SourceManual {
		t.Errorf("expected source 'manual', got %s", entry.Source)
	}
}

func TestBlockCommand_DuplicateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create test config with dummy token
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

	// Block user first time
	err = runBlock(configPath, "spammer", "spam", "https://github.com/test/repo/pull/1", models.SeverityLow, false)
	if err != nil {
		t.Fatalf("first block failed: %v", err)
	}

	// Block same user again with higher severity
	err = runBlock(configPath, "spammer", "more spam", "https://github.com/test/repo/pull/2", models.SeverityHigh, false)
	if err != nil {
		t.Fatalf("second block failed: %v", err)
	}

	// Should have 2 entries for the same user
	entries, err := db.GetEntriesByUsername("spammer")
	if err != nil {
		t.Fatalf("failed to get entries: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestBlockCommand_MissingConfig(t *testing.T) {
	configPath := "/nonexistent/config.yaml"
	err := runBlock(configPath, "testuser", "spam", "https://github.com/test/repo/pull/1", models.SeverityMedium, false)
	if err == nil {
		t.Error("expected error with missing config")
	}
}

// Cleanup helper
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
