package commands

import (
	"fmt"

	"github.com/prguard/prguard/pkg/models"
	"github.com/spf13/cobra"
)

// NewBlockCommand creates the block command
func NewBlockCommand(configPath *string) *cobra.Command {
	var reason, evidenceURL, severity string

	cmd := &cobra.Command{
		Use:   "block <username>",
		Short: "Add a user to the blocklist",
		Long:  `Blocks a GitHub user by adding them to the local blocklist`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlock(*configPath, args[0], reason, evidenceURL, severity)
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Reason for blocking (required)")
	cmd.Flags().StringVarP(&evidenceURL, "evidence", "e", "", "URL to evidence (PR/issue link, required)")
	cmd.Flags().StringVarP(&severity, "severity", "s", "medium", "Severity level (low/medium/high)")
	cmd.MarkFlagRequired("reason")
	cmd.MarkFlagRequired("evidence")

	return cmd
}

func runBlock(configPath, username, reason, evidenceURL, severity string) error {
	cfg, _, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Validate severity
	if severity != models.SeverityLow && severity != models.SeverityMedium && severity != models.SeverityHigh {
		return fmt.Errorf("invalid severity, must be low/medium/high")
	}

	// Determine who is blocking
	blockedBy := cfg.GitHub.User
	if blockedBy == "" {
		blockedBy = cfg.GitHub.Org
	}

	// Block the user
	entry, err := blManager.Block(username, reason, evidenceURL, blockedBy, severity, models.SourceManual)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	fmt.Printf("✓ User %s has been blocked\n", username)
	fmt.Printf("  ID: %s\n", entry.ID)
	fmt.Printf("  Reason: %s\n", entry.Reason)
	fmt.Printf("  Evidence: %s\n", entry.EvidenceURL)
	fmt.Printf("  Severity: %s\n", entry.Severity)

	return nil
}
