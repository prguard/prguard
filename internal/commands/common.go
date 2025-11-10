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
	if cfg.Database.Type != "sqlite" {
		return nil, fmt.Errorf("only sqlite is currently supported")
	}
	return database.NewSQLiteDB(cfg.Database.Path)
}

// initClients initializes the GitHub client and blocklist manager
func initClients(configPath string) (*config.Config, *github.Client, *blocklist.Manager, *database.DB, error) {
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
