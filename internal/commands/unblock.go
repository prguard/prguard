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
