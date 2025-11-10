package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	GitHub       GitHubConfig    `yaml:"github"`
	Database     DatabaseConfig  `yaml:"database"`
	Repositories []Repository    `yaml:"repositories"`
	Filters      FiltersConfig   `yaml:"filters"`
	Blocklist    BlocklistConfig `yaml:"blocklist"`
	Actions      ActionsConfig   `yaml:"actions"`
}

// Repository represents a GitHub repository to monitor
type Repository struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

// FullName returns the repository in owner/name format
func (r Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// GitHubConfig holds GitHub API configuration
type GitHubConfig struct {
	Token string `yaml:"token"`
	Org   string `yaml:"org"`
	User  string `yaml:"user"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type      string `yaml:"type"`       // sqlite or turso
	Path      string `yaml:"path"`       // for sqlite
	URL       string `yaml:"url"`        // for turso
	AuthToken string `yaml:"auth_token"` // for turso
}

// FiltersConfig holds PR quality filter configuration
type FiltersConfig struct {
	MinFiles        int      `yaml:"min_files"`
	MinLines        int      `yaml:"min_lines"`
	AccountAgeDays  int      `yaml:"account_age_days"`
	ReadmeOnlyBlock bool     `yaml:"readme_only_block"`
	Whitelist       []string `yaml:"whitelist"`
	SpamPhrases     []string `yaml:"spam_phrases"`
}

// BlocklistConfig holds blocklist management configuration
type BlocklistConfig struct {
	AutoExport bool              `yaml:"auto_export"`
	ExportPath string            `yaml:"export_path"`
	Sources    []BlocklistSource `yaml:"sources"`
}

// BlocklistSource represents a remote blocklist source
type BlocklistSource struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Trusted  bool   `yaml:"trusted"`
	AutoSync bool   `yaml:"auto_sync"`
}

// ActionsConfig holds default action configuration
type ActionsConfig struct {
	ClosePRs        bool   `yaml:"close_prs"`
	BlockUsers      bool   `yaml:"block_users"`
	AddSpamLabel    bool   `yaml:"add_spam_label"`
	CommentTemplate string `yaml:"comment_template"`
}

// FindConfigPath searches for a config file in standard locations
func FindConfigPath(userSpecified string) (string, error) {
	// If user specified a path, use it
	if userSpecified != "" && userSpecified != "config.yaml" {
		if _, err := os.Stat(userSpecified); err == nil {
			return userSpecified, nil
		}
		return "", fmt.Errorf("config file not found at specified path: %s", userSpecified)
	}

	// Check standard locations in order
	locations := []string{}

	// 1. ~/.config/prguard/config.yaml
	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(home, ".config", "prguard", "config.yaml"))
	}

	// 2. Current directory
	locations = append(locations, "config.yaml")

	// Return the first existing config file
	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("config file not found in any standard location: %v", locations)
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Find the actual config path
	configPath, err := FindConfigPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to a file
func Save(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the config
func applyEnvOverrides(config *Config) {
	if token := os.Getenv("PRGUARD_GITHUB_TOKEN"); token != "" {
		config.GitHub.Token = token
	}
	if org := os.Getenv("PRGUARD_GITHUB_ORG"); org != "" {
		config.GitHub.Org = org
	}
	if user := os.Getenv("PRGUARD_GITHUB_USER"); user != "" {
		config.GitHub.User = user
	}
	if dbType := os.Getenv("PRGUARD_DATABASE_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}
	if dbPath := os.Getenv("PRGUARD_DATABASE_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}
	if dbURL := os.Getenv("PRGUARD_DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}
	if authToken := os.Getenv("PRGUARD_DATABASE_AUTH_TOKEN"); authToken != "" {
		config.Database.AuthToken = authToken
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate GitHub config
	if c.GitHub.Token == "" {
		return fmt.Errorf("github.token is required")
	}
	if c.GitHub.Org == "" && c.GitHub.User == "" {
		return fmt.Errorf("either github.org or github.user must be specified")
	}

	// Validate database config
	if c.Database.Type == "" {
		return fmt.Errorf("database.type is required")
	}
	if c.Database.Type != "sqlite" && c.Database.Type != "turso" {
		return fmt.Errorf("database.type must be 'sqlite' or 'turso'")
	}
	if c.Database.Type == "sqlite" && c.Database.Path == "" {
		return fmt.Errorf("database.path is required for sqlite")
	}
	if c.Database.Type == "turso" && c.Database.URL == "" {
		return fmt.Errorf("database.url is required for turso")
	}

	return nil
}

// SetDefaults sets default values for optional configuration fields
func (c *Config) SetDefaults() {
	if c.Filters.MinFiles == 0 {
		c.Filters.MinFiles = 2
	}
	if c.Filters.MinLines == 0 {
		c.Filters.MinLines = 10
	}
	if c.Filters.AccountAgeDays == 0 {
		c.Filters.AccountAgeDays = 7
	}
	if c.Blocklist.ExportPath == "" {
		c.Blocklist.ExportPath = "./exports"
	}
	if c.Actions.CommentTemplate == "" {
		c.Actions.CommentTemplate = "This PR has been automatically closed due to low quality indicators.\nIf you believe this is an error, please contact the maintainers."
	}
	// Set default database path if using sqlite
	if c.Database.Type == "sqlite" && c.Database.Path == "" {
		if home, err := os.UserHomeDir(); err == nil {
			c.Database.Path = filepath.Join(home, ".local", "prguard", "prguard.db")
		} else {
			c.Database.Path = "./prguard.db"
		}
	}
}

// IsReadmeFile checks if a filename is a README file
func IsReadmeFile(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasPrefix(lower, "readme")
}
