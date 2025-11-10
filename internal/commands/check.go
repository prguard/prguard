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
