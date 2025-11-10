package commands

import (
	"fmt"

	"github.com/prguard/prguard/internal/blocklist"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/prguard/prguard/internal/github"
)

// loadConfig loads and validates the configuration
func loadConfig(configPath string) (*config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	cfg.SetDefaults()
	return cfg, nil
}

// initDatabase initializes the database connection
func initDatabase(cfg *config.Config) (*database.DB, error) {
	switch cfg.Database.Type {
	case "sqlite":
		return database.NewSQLiteDB(cfg.Database.Path)
	case "turso":
		return database.NewTursoDB(cfg.Database.URL, cfg.Database.AuthToken)
	default:
		return nil, fmt.Errorf("unsupported database type: %s (must be 'sqlite' or 'turso')", cfg.Database.Type)
	}
}

// initClients initializes the GitHub client and blocklist manager
func initClients(configPath string) (*config.Config, github.GitHubClient, blocklist.BlocklistManager, *database.DB, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := initDatabase(cfg)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	ghClient := github.NewClient(cfg.GitHub.Token)
	blManager := blocklist.NewManager(db)

	return cfg, ghClient, blManager, db, nil
}
