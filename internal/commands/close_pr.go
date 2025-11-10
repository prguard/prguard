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
	"strconv"

	"github.com/spf13/cobra"
)

// NewClosePRCommand creates the close-pr command
func NewClosePRCommand(configPath *string) *cobra.Command {
	var repo, comment string
	var addLabel bool

	cmd := &cobra.Command{
		Use:   "close-pr <owner>/<repo> <pr-number>...",
		Short: "Close one or more spam pull requests",
		Long:  `Closes pull requests and optionally adds a spam label`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			repo = args[0]
			prNumbers := args[1:]
			return runClosePR(*configPath, repo, prNumbers, comment, addLabel)
		},
	}

	cmd.Flags().StringVarP(&comment, "comment", "c", "", "Comment to add (uses config default if not specified)")
	cmd.Flags().BoolVarP(&addLabel, "label", "l", false, "Add 'spam' label to the PR")

	return cmd
}

func runClosePR(configPath, repo string, prNumbers []string, comment string, addLabel bool) error {
	cfg, ghClient, _, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck

	// Parse owner/repo
	owner, repoName, err := parseRepo(repo)
	if err != nil {
		return err
	}

	// Use default comment if not specified
	if comment == "" && cfg.Actions.CommentTemplate != "" {
		comment = cfg.Actions.CommentTemplate
	}

	// Close each PR
	for _, prNumStr := range prNumbers {
		prNum, err := strconv.Atoi(prNumStr)
		if err != nil {
			return fmt.Errorf("invalid PR number: %s", prNumStr)
		}

		fmt.Printf("Closing PR #%d...\n", prNum)

		// Add label if requested
		if addLabel || cfg.Actions.AddSpamLabel {
			if err := ghClient.AddLabel(owner, repoName, prNum, "spam"); err != nil {
				fmt.Printf("  Warning: failed to add label: %v\n", err)
			}
		}

		// Close the PR
		if err := ghClient.ClosePullRequest(owner, repoName, prNum, comment); err != nil {
			return fmt.Errorf("failed to close PR #%d: %w", prNum, err)
		}

		fmt.Printf("  ✓ PR #%d closed\n", prNum)
	}

	return nil
}
