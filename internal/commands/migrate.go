package commands

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/database"
	"github.com/spf13/cobra"
)

const (
	errMigrationStatus = "failed to get migration status: %w"
)

// NewMigrateCommand creates the migrate command
func NewMigrateCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
		Long:  `Manage database migrations (up, down, status)`,
	}

	cmd.AddCommand(newMigrateUpCommand(configPath))
	cmd.AddCommand(newMigrateDownCommand(configPath))
	cmd.AddCommand(newMigrateStatusCommand(configPath))

	return cmd
}

func newMigrateUpCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run pending migrations",
		Long:  `Applies all pending database migrations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateUp(*configPath)
		},
	}
}

func newMigrateDownCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Rollback last migration",
		Long:  `Rolls back the most recent database migration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateDown(*configPath)
		},
	}
}

func newMigrateStatusCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  `Displays the current database migration version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateStatus(*configPath)
		},
	}
}

func runMigrateUp(configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	cfg.SetDefaults()

	db, err := openDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Get database URL and auth token based on type
	var dbURL, authToken string
	if cfg.Database.Type == "sqlite" {
		dbURL = cfg.Database.Path
	} else {
		dbURL = cfg.Database.URL
		authToken = cfg.Database.AuthToken
	}

	if err := database.RunMigrations(db, cfg.Database.Type, dbURL, authToken); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	version, err := database.MigrationStatus(db)
	if err != nil {
		return fmt.Errorf(errMigrationStatus, err)
	}

	fmt.Printf("✓ Migrations completed successfully\n")
	fmt.Printf("  Current version: %d\n", version)

	return nil
}

func runMigrateDown(configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	cfg.SetDefaults()

	db, err := openDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Get version before rollback
	versionBefore, err := database.MigrationStatus(db)
	if err != nil {
		return fmt.Errorf(errMigrationStatus, err)
	}

	if versionBefore == 0 {
		fmt.Println("No migrations to roll back")
		return nil
	}

	// Get database URL and auth token based on type
	var dbURL, authToken string
	if cfg.Database.Type == "sqlite" {
		dbURL = cfg.Database.Path
	} else {
		dbURL = cfg.Database.URL
		authToken = cfg.Database.AuthToken
	}

	if err := database.Rollback(db, cfg.Database.Type, dbURL, authToken); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	versionAfter, err := database.MigrationStatus(db)
	if err != nil {
		return fmt.Errorf(errMigrationStatus, err)
	}

	fmt.Printf("✓ Migration rolled back successfully\n")
	fmt.Printf("  Version: %d -> %d\n", versionBefore, versionAfter)

	return nil
}

func runMigrateStatus(configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	cfg.SetDefaults()

	db, err := openDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	version, err := database.MigrationStatus(db)
	if err != nil {
		return fmt.Errorf(errMigrationStatus, err)
	}

	fmt.Printf("Database migration status\n")
	fmt.Printf("  Current version: %d\n", version)

	return nil
}

// openDatabase opens a raw database connection (not wrapped in DB struct)
func openDatabase(cfg *config.Config) (*sql.DB, error) {
	switch cfg.Database.Type {
	case "sqlite":
		// Ensure directory exists
		dir := filepath.Dir(cfg.Database.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}

		db, err := sql.Open("sqlite3", cfg.Database.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		return db, nil

	case "turso":
		connStr := cfg.Database.URL + "?authToken=" + cfg.Database.AuthToken
		db, err := sql.Open("libsql", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open turso database: %w", err)
		}

		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping turso database: %w", err)
		}

		return db, nil

	default:
		return nil, fmt.Errorf("unsupported database type: %s (must be 'sqlite' or 'turso')", cfg.Database.Type)
	}
}
