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

package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prguard/prguard/internal/blocklist"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

func TestExportCommand_JSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")
	exportPath := filepath.Join(tempDir, "export.json")

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
	defer db.Close() //nolint:errcheck

	// Add test entries
	manager := blocklist.NewManager(db)
	testUsers := []string{"user1", "user2", "user3"}
	for _, username := range testUsers {
		_, err := manager.Block(username, "spam", "https://github.com/test/repo/pull/1", "test-org", models.SeverityMedium, models.SourceManual)
		if err != nil {
			t.Fatalf("failed to block user %s: %v", username, err)
		}
	}

	// Export to JSON
	err = runExport(configPath, "json", exportPath)
	if err != nil {
		t.Errorf("runExport failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatal("export file was not created")
	}

	// Verify contents
	data, err := os.ReadFile(exportPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var entries []models.BlocklistEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(entries) != len(testUsers) {
		t.Errorf("expected %d entries in export, got %d", len(testUsers), len(entries))
	}

	// Verify usernames
	usernames := make(map[string]bool)
	for _, entry := range entries {
		usernames[entry.Username] = true
	}

	for _, username := range testUsers {
		if !usernames[username] {
			t.Errorf("expected user %s in export", username)
		}
	}
}

func TestExportCommand_CSV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")
	exportPath := filepath.Join(tempDir, "export.csv")

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
	defer db.Close() //nolint:errcheck

	// Add test entry
	manager := blocklist.NewManager(db)
	_, err = manager.Block("csvuser", "spam", "https://github.com/test/repo/pull/1", "testowner", models.SeverityHigh, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to block user: %v", err)
	}

	// Export to CSV
	err = runExport(configPath, "csv", exportPath)
	if err != nil {
		t.Errorf("runExport failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatal("export file was not created")
	}

	// Verify contents (basic check)
	data, err := os.ReadFile(exportPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	csvContent := string(data)
	if !strings.Contains(csvContent, "csvuser") {
		t.Error("CSV should contain username 'csvuser'")
	}
	if !strings.Contains(csvContent, "spam") {
		t.Error("CSV should contain reason 'spam'")
	}
}

func TestExportCommand_DefaultPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")

	// Change to temp directory so default export path is there
	oldWd, _ := os.Getwd() //nolint:errcheck
	defer os.Chdir(oldWd)  //nolint:errcheck
	_ = os.Chdir(tempDir)  //nolint:errcheck

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
	defer db.Close() //nolint:errcheck

	// Export with default path (empty string)
	err = runExport(configPath, "json", "")
	if err != nil {
		t.Errorf("runExport with default path failed: %v", err)
	}

	// Verify default file was created
	defaultPath := filepath.Join(tempDir, "blocklist.json")
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		t.Error("default export file 'blocklist.json' was not created")
	}
}

func TestExportCommand_InvalidFormat(t *testing.T) {
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
	defer db.Close() //nolint:errcheck

	// Try invalid format
	err = runExport(configPath, "xml", "")
	if err == nil {
		t.Error("expected error with invalid format")
	}
}

func TestExportCommand_EmptyBlocklist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary test directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")
	exportPath := filepath.Join(tempDir, "empty.json")

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

	// Initialize database (no entries added)
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close() //nolint:errcheck

	// Export empty blocklist
	err = runExport(configPath, "json", exportPath)
	if err != nil {
		t.Errorf("runExport with empty blocklist failed: %v", err)
	}

	// Verify file was created with empty array
	data, err := os.ReadFile(exportPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var entries []models.BlocklistEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries in empty export, got %d", len(entries))
	}
}
