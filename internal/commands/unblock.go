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

// NewUnblockCommand creates the unblock command
func NewUnblockCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unblock <username>",
		Short: "Remove a user from the blocklist",
		Long:  `Removes all blocklist entries for a GitHub user`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnblock(*configPath, args[0])
		},
	}
	return cmd
}

func runUnblock(configPath, username string) error {
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

	if !blocked {
		fmt.Printf("User %s is not in the blocklist\n", username)
		return nil
	}

	// Unblock the user
	if err := blManager.Unblock(username); err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	fmt.Printf("✓ User %s has been removed from the blocklist\n", username)
	return nil
}
