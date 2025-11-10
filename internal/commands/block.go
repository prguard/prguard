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

	"github.com/prguard/prguard/pkg/models"
	"github.com/spf13/cobra"
)

// NewBlockCommand creates the block command
func NewBlockCommand(configPath *string) *cobra.Command {
	var reason, evidenceURL, severity string
	var githubBlock bool

	cmd := &cobra.Command{
		Use:   "block <username>",
		Short: "Add a user to the blocklist",
		Long: `Blocks a GitHub user by adding them to the local blocklist.

Optionally blocks them via GitHub API using --github-block flag.
Note: GitHub blocking works at organization or personal account level, not per-repository.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runBlock(*configPath, args[0], reason, evidenceURL, severity, githubBlock)
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Reason for blocking (required)")
	cmd.Flags().StringVarP(&evidenceURL, "evidence", "e", "", "URL to evidence (PR/issue link, required)")
	cmd.Flags().StringVarP(&severity, "severity", "s", "medium", "Severity level (low/medium/high)")
	cmd.Flags().BoolVar(&githubBlock, "github-block", false, "Also block user via GitHub API (affects ALL repos in org/account)")
	_ = cmd.MarkFlagRequired("reason")
	_ = cmd.MarkFlagRequired("evidence")

	return cmd
}

func runBlock(configPath, username, reason, evidenceURL, severity string, githubBlock bool) error {
	cfg, ghClient, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck

	// Validate severity
	if severity != models.SeverityLow && severity != models.SeverityMedium && severity != models.SeverityHigh {
		return fmt.Errorf("invalid severity, must be low/medium/high")
	}

	// Determine who is blocking
	blockedBy := cfg.GitHub.User
	if blockedBy == "" {
		blockedBy = cfg.GitHub.Org
	}

	// Add to local blocklist
	entry, err := blManager.Block(username, reason, evidenceURL, blockedBy, severity, models.SourceManual)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	fmt.Printf("✓ User %s added to local blocklist\n", username)
	fmt.Printf("  ID: %s\n", entry.ID)
	fmt.Printf("  Reason: %s\n", entry.Reason)
	fmt.Printf("  Evidence: %s\n", entry.EvidenceURL)
	fmt.Printf("  Severity: %s\n", entry.Severity)

	// GitHub API blocking (optional)
	if githubBlock {
		fmt.Println()
		switch {
		case cfg.GitHub.Org != "":
			// Organization-level blocking
			fmt.Printf("⚠️  WARNING: This will block %s from ALL repositories in the '%s' organization.\n", username, cfg.GitHub.Org)
			fmt.Print("Continue? (y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("GitHub blocking cancelled. User remains in local blocklist.")
				return nil
			}

			if err := ghClient.BlockUserOrg(cfg.GitHub.Org, username); err != nil {
				return fmt.Errorf("failed to block user via GitHub API: %w", err)
			}
			fmt.Printf("✓ User %s blocked at organization level via GitHub API\n", username)
			fmt.Printf("  Scope: ALL repositories in '%s' organization\n", cfg.GitHub.Org)
			fmt.Println("  Required permission: admin:org")
		case cfg.GitHub.User != "":
			// Personal account-level blocking
			fmt.Printf("⚠️  WARNING: This will block %s from ALL repositories owned by your personal account (%s).\n", username, cfg.GitHub.User)
			fmt.Print("Continue? (y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("GitHub blocking cancelled. User remains in local blocklist.")
				return nil
			}

			if err := ghClient.BlockUserPersonal(username); err != nil {
				return fmt.Errorf("failed to block user via GitHub API: %w", err)
			}
			fmt.Printf("✓ User %s blocked at personal account level via GitHub API\n", username)
			fmt.Printf("  Scope: ALL repositories owned by '%s'\n", cfg.GitHub.User)
			fmt.Println("  Required permission: user")
		default:
			return fmt.Errorf("cannot use --github-block: neither github.org nor github.user is configured")
		}
	} else {
		fmt.Println("\nNote: User is only blocked in PRGuard's local database.")
		fmt.Println("To also block via GitHub API, use: --github-block flag")
		fmt.Println("(This will block them from ALL repos in your org/account)")
	}

	return nil
}
