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
	"fmt"

	"github.com/spf13/cobra"
)

// NewScanAllCommand creates the scan-all command
func NewScanAllCommand(configPath *string) *cobra.Command {
	var autoClose, autoBlock, githubBlock bool

	cmd := &cobra.Command{
		Use:   "scan-all",
		Short: "Scan all configured repositories for spam pull requests",
		Long: `Scans all repositories listed in the configuration file for spam indicators.

By default, scan-all only reports findings. Use flags to take action:
  --auto-close: Automatically close spam PRs
  --auto-block: Automatically add spam users to local blocklist
  --github-block: Also block users via GitHub API (requires --auto-block)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScanAll(*configPath, autoClose, autoBlock, githubBlock)
		},
	}

	cmd.Flags().BoolVar(&autoClose, "auto-close", false, "Automatically close spam PRs")
	cmd.Flags().BoolVar(&autoBlock, "auto-block", false, "Automatically add spam users to blocklist")
	cmd.Flags().BoolVar(&githubBlock, "github-block", false, "Also block users via GitHub API (requires --auto-block)")

	return cmd
}

func runScanAll(configPath string, autoClose, autoBlock, githubBlock bool) error {
	// Validate flags
	if githubBlock && !autoBlock {
		return fmt.Errorf("--github-block requires --auto-block")
	}

	cfg, _, _, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Add repositories to your config.yaml file")
	}

	fmt.Printf("Scanning %d configured repositories...\n\n", len(cfg.Repositories))

	for _, repo := range cfg.Repositories {
		fmt.Printf("=== %s ===\n", repo.FullName())

		// Run individual scan for each repository
		if err := runScan(configPath, repo.FullName(), autoClose, autoBlock, githubBlock); err != nil {
			fmt.Printf("Error scanning %s: %v\n\n", repo.FullName(), err)
			continue
		}

		fmt.Println()
	}

	return nil
}
