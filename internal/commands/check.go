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

// NewCheckCommand creates the check command
func NewCheckCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <username>",
		Short: "Check if a user is in the blocklist",
		Long:  `Checks if a GitHub user is currently blocked`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(*configPath, args[0])
		},
	}
	return cmd
}

func runCheck(configPath, username string) error {
	_, _, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check if user is blocked
	blocked, err := blManager.IsBlocked(username)
	if err != nil {
		return fmt.Errorf("failed to check block status: %w", err)
	}

	if blocked {
		fmt.Printf("✓ User %s is BLOCKED\n\n", username)

		// Get entries
		entries, err := blManager.GetByUsername(username)
		if err != nil {
			return fmt.Errorf("failed to get entries: %w", err)
		}

		for _, entry := range entries {
			fmt.Printf("Entry ID: %s\n", entry.ID)
			fmt.Printf("  Reason: %s\n", entry.Reason)
			fmt.Printf("  Evidence: %s\n", entry.EvidenceURL)
			fmt.Printf("  Severity: %s\n", entry.Severity)
			fmt.Printf("  Blocked by: %s\n", entry.BlockedBy)
			fmt.Printf("  Source: %s\n", entry.Source)
			fmt.Printf("  Date: %s\n\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Printf("User %s is NOT blocked\n", username)
	}

	return nil
}
