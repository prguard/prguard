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

package database

import (
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// checkGeniInstalled checks if geni CLI is installed
func checkGeniInstalled() error {
	_, err := exec.LookPath("geni")
	if err != nil {
		return fmt.Errorf(`geni CLI not found. Please install it:
  - Homebrew: brew install geni
  - Cargo:    cargo install geni
  - See:      https://github.com/emilpriver/geni`)
	}
	return nil
}

// getMigrationsPath returns the absolute path to the migrations directory
func getMigrationsPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "migrations"), nil
}

// RunMigrations executes all pending migrations using geni CLI
// dbType: "sqlite" or "turso"
// dbURL: file path for sqlite, or libsql:// URL for turso
// authToken: empty for sqlite, auth token for turso
func RunMigrations(db *sql.DB, dbType, dbURL, authToken string) error {
	if err := checkGeniInstalled(); err != nil {
		return err
	}

	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return err
	}

	// Build DATABASE_URL based on type
	var databaseURL string
	switch dbType {
	case "sqlite":
		databaseURL = "sqlite://" + dbURL
	case "turso":
		// Convert libsql:// to https:// for geni
		databaseURL = strings.Replace(dbURL, "libsql://", "https://", 1)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Run geni up command with DATABASE_URL environment variable
	cmd := exec.Command("geni", "up")
	cmd.Env = append(cmd.Env, "DATABASE_URL="+databaseURL)
	cmd.Env = append(cmd.Env, "MIGRATIONS_DIR="+migrationsPath)
	cmd.Dir = filepath.Dir(migrationsPath)

	// Add auth token for Turso (geni uses DATABASE_TOKEN)
	if dbType == "turso" && authToken != "" {
		cmd.Env = append(cmd.Env, "DATABASE_TOKEN="+authToken)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("migration failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Rollback rolls back the last migration using geni CLI
func Rollback(db *sql.DB, dbType, dbURL, authToken string) error {
	if err := checkGeniInstalled(); err != nil {
		return err
	}

	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return err
	}

	// Build DATABASE_URL based on type
	var databaseURL string
	switch dbType {
	case "sqlite":
		databaseURL = "sqlite://" + dbURL
	case "turso":
		// Convert libsql:// to https:// for geni
		databaseURL = strings.Replace(dbURL, "libsql://", "https://", 1)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Run geni down command with DATABASE_URL environment variable
	cmd := exec.Command("geni", "down")
	cmd.Env = append(cmd.Env, "DATABASE_URL="+databaseURL)
	cmd.Env = append(cmd.Env, "MIGRATIONS_DIR="+migrationsPath)
	cmd.Dir = filepath.Dir(migrationsPath)

	// Add auth token for Turso (geni uses DATABASE_TOKEN)
	if dbType == "turso" && authToken != "" {
		cmd.Env = append(cmd.Env, "DATABASE_TOKEN="+authToken)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rollback failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// MigrationStatus returns the current migration version
func MigrationStatus(db *sql.DB) (int, error) {
	// Query the geni migrations table directly
	var version int
	err := db.QueryRow("SELECT version FROM geni_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		// Table might not exist yet
		var tableCount int
		err2 := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='geni_migrations'").Scan(&tableCount)
		if err2 != nil {
			return 0, err2
		}
		if tableCount == 0 {
			return 0, nil
		}
		return 0, err
	}

	// Parse version from filename format (001_initial_schema)
	versionStr := strconv.Itoa(version)
	if strings.HasPrefix(versionStr, "00") {
		versionStr = strings.TrimLeft(versionStr, "0")
		if versionStr == "" {
			versionStr = "0"
		}
		version, _ = strconv.Atoi(versionStr)
	}

	return version, nil
}
