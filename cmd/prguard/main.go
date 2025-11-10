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

package main

import (
	"fmt"
	"os"

	"github.com/prguard/prguard/internal/commands"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "prguard",
		Short: "PRGuard - Detect and block spam pull requests on GitHub",
		Long: `PRGuard helps open source maintainers detect, block, and manage spam pull requests.
It analyzes PR quality using configurable heuristics and maintains a portable blocklist.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Global flags
	var configPath string
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to configuration file")

	// Add commands
	rootCmd.AddCommand(commands.NewInitCommand(&configPath))
	rootCmd.AddCommand(commands.NewMigrateCommand(&configPath))
	rootCmd.AddCommand(commands.NewScanCommand(&configPath))
	rootCmd.AddCommand(commands.NewScanAllCommand(&configPath))
	rootCmd.AddCommand(commands.NewBlockCommand(&configPath))
	rootCmd.AddCommand(commands.NewUnblockCommand(&configPath))
	rootCmd.AddCommand(commands.NewCheckCommand(&configPath))
	rootCmd.AddCommand(commands.NewListCommand(&configPath))
	rootCmd.AddCommand(commands.NewExportCommand(&configPath))
	rootCmd.AddCommand(commands.NewImportCommand(&configPath))
	rootCmd.AddCommand(commands.NewClosePRCommand(&configPath))
	rootCmd.AddCommand(commands.NewReviewCommand(&configPath))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
