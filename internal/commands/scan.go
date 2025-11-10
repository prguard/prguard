package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/prguard/prguard/internal/scanner"
	"github.com/prguard/prguard/pkg/models"
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

	cfg, ghClient, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Apply config defaults if flags weren't explicitly set
	// Note: Cobra doesn't provide a way to detect if a bool flag was explicitly set,
	// so we assume false means "use config default" and true means "force enable"
	// This means CLI can only enable, not disable config defaults
	if !autoClose && cfg.Actions.ClosePRs {
		autoClose = true
	}
	if !autoBlock && cfg.Actions.BlockUsers {
		autoBlock = true
	}

	// Parse owner/repo
	owner, repoName, err := parseRepo(repo)
	if err != nil {
		return err
	}

	fmt.Printf("Scanning repository %s/%s...\n\n", owner, repoName)

	// Create scanner
	scan := scanner.NewScanner(cfg)

	// Scan repository
	results, err := scan.ScanRepository(ghClient, owner, repoName)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Display results summary
	fmt.Printf("Total PRs: %d\n", results.Total)
	fmt.Printf("Spam detected: %d\n", len(results.Spam))
	fmt.Printf("Uncertain: %d\n", len(results.Uncertain))
	fmt.Printf("Clean: %d\n\n", len(results.Clean))

	// Collect unique spam users
	spamUsers := make(map[string]struct {
		firstPR     int
		evidenceURL string
		severity    string
		reasons     []string
	})

	if len(results.Spam) > 0 {
		fmt.Println("=== SPAM DETECTED ===")
		for _, result := range results.Spam {
			fmt.Printf("\nPR #%d: %s\n", result.PR.Number, result.PR.Title)
			fmt.Printf("  Author: %s\n", result.PR.Author)
			fmt.Printf("  URL: %s\n", result.PR.HTMLURL)
			fmt.Printf("  Severity: %s\n", result.Severity)
			fmt.Printf("  Reasons:\n")
			for _, reason := range result.Reasons {
				fmt.Printf("    - %s\n", reason)
			}
			fmt.Printf("  Recommended action: %s\n", result.RecommendAction)

			// Track user for potential blocking
			if _, exists := spamUsers[result.PR.Author]; !exists {
				spamUsers[result.PR.Author] = struct {
					firstPR     int
					evidenceURL string
					severity    string
					reasons     []string
				}{
					firstPR:     result.PR.Number,
					evidenceURL: result.PR.HTMLURL,
					severity:    result.Severity,
					reasons:     result.Reasons,
				}
			}
		}
	}

	if len(results.Uncertain) > 0 {
		fmt.Println("\n=== MANUAL REVIEW NEEDED ===")
		for _, result := range results.Uncertain {
			fmt.Printf("\nPR #%d: %s\n", result.PR.Number, result.PR.Title)
			fmt.Printf("  Author: %s\n", result.PR.Author)
			fmt.Printf("  URL: %s\n", result.PR.HTMLURL)
			fmt.Printf("  Reasons:\n")
			for _, reason := range result.Reasons {
				fmt.Printf("    - %s\n", reason)
			}
		}
	}

	// Take automated actions if flags are set
	if len(results.Spam) > 0 && (autoClose || autoBlock) {
		fmt.Println("\n=== AUTOMATED ACTIONS ===")

		// Confirm with user
		if !confirmAction(len(results.Spam), len(spamUsers), autoClose, autoBlock, githubBlock) {
			fmt.Println("Actions cancelled by user.")
			return nil
		}

		// Block users first
		if autoBlock {
			fmt.Printf("\nBlocking %d spam users...\n", len(spamUsers))

			blockedBy := cfg.GitHub.User
			if blockedBy == "" {
				blockedBy = cfg.GitHub.Org
			}

			for username, info := range spamUsers {
				// Add to local blocklist
				reason := fmt.Sprintf("Auto-detected spam: %s", strings.Join(info.reasons, ", "))
				_, err := blManager.Block(username, reason, info.evidenceURL, blockedBy, info.severity, models.SourceAutoDetected)
				if err != nil {
					fmt.Printf("  ✗ Failed to block %s: %v\n", username, err)
					continue
				}
				fmt.Printf("  ✓ Blocked %s in local blocklist\n", username)

				// Block on GitHub if requested
				if githubBlock {
					if cfg.GitHub.Org != "" {
						if err := ghClient.BlockUserOrg(cfg.GitHub.Org, username); err != nil {
							fmt.Printf("    ⚠ Failed to block on GitHub (org): %v\n", err)
						} else {
							fmt.Printf("    ✓ Blocked on GitHub (org level)\n")
						}
					} else if cfg.GitHub.User != "" {
						if err := ghClient.BlockUserPersonal(username); err != nil {
							fmt.Printf("    ⚠ Failed to block on GitHub (personal): %v\n", err)
						} else {
							fmt.Printf("    ✓ Blocked on GitHub (personal level)\n")
						}
					}
				}
			}
		}

		// Close PRs
		if autoClose {
			fmt.Printf("\nClosing %d spam PRs...\n", len(results.Spam))
			comment := cfg.Actions.CommentTemplate
			if comment == "" {
				comment = "This PR has been automatically closed due to spam indicators."
			}

			for _, result := range results.Spam {
				// Add label if configured
				if cfg.Actions.AddSpamLabel {
					if err := ghClient.AddLabel(owner, repoName, result.PR.Number, "spam"); err != nil {
						fmt.Printf("  ⚠ PR #%d: failed to add label: %v\n", result.PR.Number, err)
					}
				}

				// Close the PR
				if err := ghClient.ClosePullRequest(owner, repoName, result.PR.Number, comment); err != nil {
					fmt.Printf("  ✗ PR #%d: failed to close: %v\n", result.PR.Number, err)
				} else {
					fmt.Printf("  ✓ PR #%d closed\n", result.PR.Number)
				}
			}
		}

		fmt.Println("\n✓ Automated actions completed")
	}

	// Print summary
	if len(results.Spam) > 0 && !autoClose && !autoBlock {
		fmt.Println("\nTo take action automatically, use:")
		fmt.Printf("  prguard scan %s --auto-close --auto-block\n", repo)
		if githubBlock {
			fmt.Printf("  Add --github-block to also block on GitHub\n")
		}
	}

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
