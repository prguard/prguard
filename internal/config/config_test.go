package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
github:
  token: "test-token-123"
  org: "test-org"

database:
  type: "sqlite"
  path: "/tmp/test.db"

filters:
  min_files: 3
  min_lines: 15
  account_age_days: 10
  readme_only_block: true
  whitelist:
    - "bot1"
    - "bot2"
  spam_phrases:
    - "spam1"
    - "spam2"

blocklist:
  auto_export: true
  export_path: "/tmp/exports"

actions:
  close_prs: true
  add_spam_label: true
  comment_template: "Test comment"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Validate GitHub config
	if cfg.GitHub.Token != "test-token-123" {
		t.Errorf("Expected token 'test-token-123', got '%s'", cfg.GitHub.Token)
	}
	if cfg.GitHub.Org != "test-org" {
		t.Errorf("Expected org 'test-org', got '%s'", cfg.GitHub.Org)
	}

	// Validate database config
	if cfg.Database.Type != "sqlite" {
		t.Errorf("Expected database type 'sqlite', got '%s'", cfg.Database.Type)
	}
	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("Expected database path '/tmp/test.db', got '%s'", cfg.Database.Path)
	}

	// Validate filters
	if cfg.Filters.MinFiles != 3 {
		t.Errorf("Expected min_files 3, got %d", cfg.Filters.MinFiles)
	}
	if cfg.Filters.MinLines != 15 {
		t.Errorf("Expected min_lines 15, got %d", cfg.Filters.MinLines)
	}
	if cfg.Filters.AccountAgeDays != 10 {
		t.Errorf("Expected account_age_days 10, got %d", cfg.Filters.AccountAgeDays)
	}
	if !cfg.Filters.ReadmeOnlyBlock {
		t.Error("Expected readme_only_block to be true")
	}
	if len(cfg.Filters.Whitelist) != 2 {
		t.Errorf("Expected 2 whitelist entries, got %d", len(cfg.Filters.Whitelist))
	}
	if len(cfg.Filters.SpamPhrases) != 2 {
		t.Errorf("Expected 2 spam phrases, got %d", len(cfg.Filters.SpamPhrases))
	}

	// Validate blocklist config
	if !cfg.Blocklist.AutoExport {
		t.Error("Expected auto_export to be true")
	}
	if cfg.Blocklist.ExportPath != "/tmp/exports" {
		t.Errorf("Expected export_path '/tmp/exports', got '%s'", cfg.Blocklist.ExportPath)
	}

	// Validate actions
	if !cfg.Actions.ClosePRs {
		t.Error("Expected close_prs to be true")
	}
	if !cfg.Actions.AddSpamLabel {
		t.Error("Expected add_spam_label to be true")
	}
	if cfg.Actions.CommentTemplate != "Test comment" {
		t.Errorf("Expected comment template 'Test comment', got '%s'", cfg.Actions.CommentTemplate)
	}
}

func TestLoadWithRepositories(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
github:
  token: "test-token"
  org: "test-org"

database:
  type: "sqlite"
  path: "/tmp/test.db"

repositories:
  - owner: "org1"
    name: "repo1"
  - owner: "org2"
    name: "repo2"
`

	os.WriteFile(configPath, []byte(configContent), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Repositories) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(cfg.Repositories))
	}

	if cfg.Repositories[0].Owner != "org1" || cfg.Repositories[0].Name != "repo1" {
		t.Errorf("First repository incorrect: %s/%s", cfg.Repositories[0].Owner, cfg.Repositories[0].Name)
	}

	// Test FullName method
	fullName := cfg.Repositories[0].FullName()
	if fullName != "org1/repo1" {
		t.Errorf("Expected 'org1/repo1', got '%s'", fullName)
	}
}

func TestValidate_MissingToken(t *testing.T) {
	cfg := &Config{
		GitHub: GitHubConfig{
			Token: "", // Missing token
			Org:   "test-org",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "/tmp/test.db",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing token")
	}
}

func TestValidate_MissingOrgAndUser(t *testing.T) {
	cfg := &Config{
		GitHub: GitHubConfig{
			Token: "test-token",
			Org:   "", // Missing org
			User:  "", // Missing user
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "/tmp/test.db",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing org/user")
	}
}

func TestValidate_InvalidDatabaseType(t *testing.T) {
	cfg := &Config{
		GitHub: GitHubConfig{
			Token: "test-token",
			Org:   "test-org",
		},
		Database: DatabaseConfig{
			Type: "invalid", // Invalid type
			Path: "/tmp/test.db",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid database type")
	}
}

func TestValidate_MissingDatabasePath(t *testing.T) {
	cfg := &Config{
		GitHub: GitHubConfig{
			Token: "test-token",
			Org:   "test-org",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "", // Missing path
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing database path")
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &Config{
		GitHub: GitHubConfig{
			Token: "test-token",
			Org:   "test-org",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "/tmp/test.db",
		},
		Filters:   FiltersConfig{},
		Blocklist: BlocklistConfig{},
		Actions:   ActionsConfig{},
	}

	cfg.SetDefaults()

	// Check filter defaults
	if cfg.Filters.MinFiles != 2 {
		t.Errorf("Expected default min_files 2, got %d", cfg.Filters.MinFiles)
	}
	if cfg.Filters.MinLines != 10 {
		t.Errorf("Expected default min_lines 10, got %d", cfg.Filters.MinLines)
	}
	if cfg.Filters.AccountAgeDays != 7 {
		t.Errorf("Expected default account_age_days 7, got %d", cfg.Filters.AccountAgeDays)
	}

	// Check blocklist defaults
	if cfg.Blocklist.ExportPath != "./exports" {
		t.Errorf("Expected default export_path './exports', got '%s'", cfg.Blocklist.ExportPath)
	}

	// Check actions defaults
	if cfg.Actions.CommentTemplate == "" {
		t.Error("Expected default comment template to be set")
	}
}

func TestEnvOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("PRGUARD_GITHUB_TOKEN", "env-token")
	os.Setenv("PRGUARD_GITHUB_ORG", "env-org")
	os.Setenv("PRGUARD_DATABASE_PATH", "/env/path/db")
	defer func() {
		os.Unsetenv("PRGUARD_GITHUB_TOKEN")
		os.Unsetenv("PRGUARD_GITHUB_ORG")
		os.Unsetenv("PRGUARD_DATABASE_PATH")
	}()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
github:
  token: "config-token"
  org: "config-org"

database:
  type: "sqlite"
  path: "/config/path/db"
`

	os.WriteFile(configPath, []byte(configContent), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env vars should override config file
	if cfg.GitHub.Token != "env-token" {
		t.Errorf("Expected token from env 'env-token', got '%s'", cfg.GitHub.Token)
	}
	if cfg.GitHub.Org != "env-org" {
		t.Errorf("Expected org from env 'env-org', got '%s'", cfg.GitHub.Org)
	}
	if cfg.Database.Path != "/env/path/db" {
		t.Errorf("Expected path from env '/env/path/db', got '%s'", cfg.Database.Path)
	}
}

func TestIsReadmeFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"README.md", true},
		{"readme.txt", true},
		{"README", true},
		{"Readme.rst", true},
		{"docs/README.md", false}, // Only checks basename
		{"CONTRIBUTING.md", false},
		{"main.go", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsReadmeFile(tt.filename)
			if result != tt.expected {
				t.Errorf("IsReadmeFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestFindConfigPath_CustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom-config.yaml")

	// Create the file
	os.WriteFile(customPath, []byte("test: value"), 0644)

	path, err := FindConfigPath(customPath)
	if err != nil {
		t.Fatalf("FindConfigPath failed: %v", err)
	}

	if path != customPath {
		t.Errorf("Expected custom path '%s', got '%s'", customPath, path)
	}
}

func TestFindConfigPath_NotFound(t *testing.T) {
	// Use a non-existent custom path
	_, err := FindConfigPath("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}
