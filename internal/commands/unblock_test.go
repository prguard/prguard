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
	"path/filepath"
	"testing"

	"github.com/prguard/prguard/internal/blocklist"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/pkg/models"
)

func TestUnblockCommand_Success(t *testing.T) {
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
	_, err = manager.Block("testspammer", "spam", "https://github.com/test/repo/pull/1", "test-org", models.SeverityMedium, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to block user: %v", err)
	}

	// Verify user is blocked
	blocked, err := manager.IsBlocked("testspammer")
	if err != nil {
		t.Fatalf("failed to check block status: %v", err)
	}
	if !blocked {
		t.Fatal("user should be blocked")
	}

	// Unblock user
	err = runUnblock(configPath, "testspammer")
	if err != nil {
		t.Errorf("runUnblock failed: %v", err)
	}

	// Verify user is no longer blocked
	blocked, err = manager.IsBlocked("testspammer")
	if err != nil {
		t.Fatalf("failed to check block status after unblock: %v", err)
	}
	if blocked {
		t.Error("user should not be blocked after unblock")
	}
}

func TestUnblockCommand_UserNotBlocked(t *testing.T) {
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

	// Try to unblock a user that isn't blocked - should succeed without error
	err = runUnblock(configPath, "notblocked")
	if err != nil {
		t.Errorf("runUnblock should succeed for non-blocked user: %v", err)
	}
}

func TestUnblockCommand_MultipleEntries(t *testing.T) {
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
	_, err = manager.Block("multientry", "spam 1", "https://github.com/test/repo/pull/1", "test-org", models.SeverityLow, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to block user first time: %v", err)
	}

	_, err = manager.Block("multientry", "spam 2", "https://github.com/test/repo/pull/2", "test-org", models.SeverityHigh, models.SourceManual)
	if err != nil {
		t.Fatalf("failed to block user second time: %v", err)
	}

	// Verify user has 2 entries
	entries, err := manager.GetByUsername("multientry")
	if err != nil {
		t.Fatalf("failed to get entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries before unblock, got %d", len(entries))
	}

	// Unblock should remove ALL entries for the user
	err = runUnblock(configPath, "multientry")
	if err != nil {
		t.Errorf("runUnblock failed: %v", err)
	}

	// Verify all entries are removed
	blocked, err := manager.IsBlocked("multientry")
	if err != nil {
		t.Fatalf("failed to check block status: %v", err)
	}
	if blocked {
		t.Error("user should not be blocked after unblock")
	}

	entries, err = manager.GetByUsername("multientry")
	if err != nil {
		t.Fatalf("failed to get entries after unblock: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after unblock, got %d", len(entries))
	}
}

func TestUnblockCommand_MissingConfig(t *testing.T) {
	configPath := "/nonexistent/config.yaml"
	err := runUnblock(configPath, "testuser")
	if err == nil {
		t.Error("expected error with missing config")
	}
}
