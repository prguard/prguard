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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/prguard/prguard/internal/scanner"
	"github.com/spf13/cobra"
)

// NewScanCommand creates the scan command
func NewScanCommand(configPath *string) *cobra.Command {
	var autoClose, autoBlock, githubBlock bool

	cmd := &cobra.Command{
		Use:   "scan <owner>/<repo>",
		Short: "Scan a repository for spam pull requests",
		Long: `Analyzes all open pull requests in a repository for spam indicators.

By default, scan only reports findings. Use flags to take action:
  --auto-close: Automatically close spam PRs
  --auto-block: Automatically add spam users to local blocklist
  --github-block: Also block users via GitHub API (requires --auto-block)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(*configPath, args[0], autoClose, autoBlock, githubBlock)
		},
	}

	cmd.Flags().BoolVar(&autoClose, "auto-close", false, "Automatically close spam PRs")
	cmd.Flags().BoolVar(&autoBlock, "auto-block", false, "Automatically add spam users to blocklist")
	cmd.Flags().BoolVar(&githubBlock, "github-block", false, "Also block users via GitHub API (requires --auto-block)")

	return cmd
}

func runScan(configPath, repo string, autoClose, autoBlock, githubBlock bool) error {
	// Validate flags
	if githubBlock && !autoBlock {
		return fmt.Errorf("--github-block requires --auto-block")
	}

	// Initialize clients and database
	cfg, ghClient, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Apply config defaults to flags
	autoClose, autoBlock = applyConfigDefaults(cfg, autoClose, autoBlock)

	// Parse owner/repo
	owner, repoName, err := parseRepo(repo)
	if err != nil {
		return err
	}

	fmt.Printf("Scanning repository %s/%s...\n\n", owner, repoName)

	// Scan repository for spam PRs
	scan := scanner.NewScanner(cfg)
	results, err := scan.ScanRepository(ghClient, owner, repoName)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Display scan results
	displayScanSummary(results)
	spamUsers := collectSpamUsers(results)
	displayUncertainResults(results)

	// Execute automated actions if requested
	ctx := &ActionContext{
		cfg:       cfg,
		ghClient:  ghClient,
		blManager: blManager,
	}
	flags := &ActionFlags{
		autoClose:   autoClose,
		autoBlock:   autoBlock,
		githubBlock: githubBlock,
	}
	if err := executeAutomatedActions(ctx, owner, repoName, results, spamUsers, flags); err != nil {
		return err
	}

	// Show suggestions if no actions taken
	displayActionSuggestions(repo, len(results.Spam) > 0, autoClose, autoBlock, githubBlock)

	return nil
}

func confirmAction(numPRs, numUsers int, autoClose, autoBlock, githubBlock bool) bool {
	fmt.Println()
	fmt.Printf("About to take the following actions:\n")
	if autoBlock {
		fmt.Printf("  - Add %d users to local blocklist\n", numUsers)
		if githubBlock {
			fmt.Printf("  - Block %d users via GitHub API (ALL repos)\n", numUsers)
		}
	}
	if autoClose {
		fmt.Printf("  - Close %d spam PRs\n", numPRs)
	}

	fmt.Print("\nContinue? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

func parseRepo(repo string) (owner, name string, err error) {
	// Simple parser for owner/repo format
	for i, c := range repo {
		if c == '/' {
			return repo[:i], repo[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid repository format, expected owner/repo")
}
