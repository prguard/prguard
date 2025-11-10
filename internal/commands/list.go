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

// NewListCommand creates the list command
func NewListCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all blocklist entries",
		Long:  `Displays all users in the blocklist with their details`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(*configPath)
		},
	}
	return cmd
}

func runList(configPath string) error {
	_, _, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	entries, err := blManager.List()
	if err != nil {
		return fmt.Errorf("failed to list entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("Blocklist is empty")
		return nil
	}

	fmt.Printf("Total blocked users: %d\n\n", len(entries))

	for i, entry := range entries {
		fmt.Printf("%d. %s\n", i+1, entry.Username)
		fmt.Printf("   ID: %s\n", entry.ID)
		fmt.Printf("   Reason: %s\n", entry.Reason)
		fmt.Printf("   Evidence: %s\n", entry.EvidenceURL)
		fmt.Printf("   Severity: %s\n", entry.Severity)
		fmt.Printf("   Blocked by: %s\n", entry.BlockedBy)
		fmt.Printf("   Source: %s\n", entry.Source)
		fmt.Printf("   Date: %s\n\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
	}

	return nil
}
