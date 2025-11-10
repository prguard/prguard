package commands

import (
	"fmt"
	"strings"

	"github.com/prguard/prguard/internal/blocklist"
	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/github"
	"github.com/prguard/prguard/internal/scanner"
	"github.com/prguard/prguard/pkg/models"
)

// spamUserInfo holds information about a spam user
type spamUserInfo struct {
	firstPR     int
	evidenceURL string
	severity    string
	reasons     []string
}

// ActionContext holds service dependencies for executing actions
type ActionContext struct {
	cfg       *config.Config
	ghClient  *github.Client
	blManager *blocklist.Manager
}

// ActionFlags holds configuration for which actions to execute
type ActionFlags struct {
	autoClose   bool
	autoBlock   bool
	githubBlock bool
}

// applyConfigDefaults applies config defaults to action flags
func applyConfigDefaults(cfg *config.Config, autoClose, autoBlock bool) (bool, bool) {
	// Apply config defaults if flags weren't explicitly set
	// Note: Cobra doesn't provide a way to detect if a bool flag was explicitly set,
	// so we assume false means "use config default" and true means "force enable"
	if !autoClose && cfg.Actions.ClosePRs {
		autoClose = true
	}
	if !autoBlock && cfg.Actions.BlockUsers {
		autoBlock = true
	}
	return autoClose, autoBlock
}

// displayScanSummary prints the scan results summary
func displayScanSummary(results *scanner.ScanResults) {
	fmt.Printf("Total PRs: %d\n", results.Total)
	fmt.Printf("Spam detected: %d\n", len(results.Spam))
	fmt.Printf("Uncertain: %d\n", len(results.Uncertain))
	fmt.Printf("Clean: %d\n\n", len(results.Clean))
}

// collectSpamUsers collects unique spam users and displays spam PRs
func collectSpamUsers(results *scanner.ScanResults) map[string]spamUserInfo {
	spamUsers := make(map[string]spamUserInfo)

	if len(results.Spam) == 0 {
		return spamUsers
	}

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
			spamUsers[result.PR.Author] = spamUserInfo{
				firstPR:     result.PR.Number,
				evidenceURL: result.PR.HTMLURL,
				severity:    result.Severity,
				reasons:     result.Reasons,
			}
		}
	}

	return spamUsers
}

// displayUncertainResults shows PRs that need manual review
func displayUncertainResults(results *scanner.ScanResults) {
	if len(results.Uncertain) == 0 {
		return
	}

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

// executeBlockActions blocks spam users in local blocklist and optionally on GitHub
func executeBlockActions(ctx *ActionContext, spamUsers map[string]spamUserInfo, githubBlock bool) {
	fmt.Printf("\nBlocking %d spam users...\n", len(spamUsers))

	blockedBy := ctx.cfg.GitHub.User
	if blockedBy == "" {
		blockedBy = ctx.cfg.GitHub.Org
	}

	for username, info := range spamUsers {
		// Add to local blocklist
		reason := fmt.Sprintf("Auto-detected spam: %s", strings.Join(info.reasons, ", "))
		_, err := ctx.blManager.Block(username, reason, info.evidenceURL, blockedBy, info.severity, models.SourceAutoDetected)
		if err != nil {
			fmt.Printf("  ✗ Failed to block %s: %v\n", username, err)
			continue
		}
		fmt.Printf("  ✓ Blocked %s in local blocklist\n", username)

		// Block on GitHub if requested
		if githubBlock {
			blockOnGitHub(ctx.cfg, ctx.ghClient, username)
		}
	}
}

// blockOnGitHub blocks a user via GitHub API (org or personal)
func blockOnGitHub(cfg *config.Config, ghClient *github.Client, username string) {
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

// executeCloseActions closes spam PRs and optionally adds labels
func executeCloseActions(ctx *ActionContext, owner, repoName string, results *scanner.ScanResults) {
	fmt.Printf("\nClosing %d spam PRs...\n", len(results.Spam))

	comment := ctx.cfg.Actions.CommentTemplate
	if comment == "" {
		comment = "This PR has been automatically closed due to spam indicators."
	}

	for _, result := range results.Spam {
		// Add label if configured
		if ctx.cfg.Actions.AddSpamLabel {
			if err := ctx.ghClient.AddLabel(owner, repoName, result.PR.Number, "spam"); err != nil {
				fmt.Printf("  ⚠ PR #%d: failed to add label: %v\n", result.PR.Number, err)
			}
		}

		// Close the PR
		if err := ctx.ghClient.ClosePullRequest(owner, repoName, result.PR.Number, comment); err != nil {
			fmt.Printf("  ✗ PR #%d: failed to close: %v\n", result.PR.Number, err)
		} else {
			fmt.Printf("  ✓ PR #%d closed\n", result.PR.Number)
		}
	}
}

// executeAutomatedActions orchestrates blocking and closing actions
func executeAutomatedActions(
	ctx *ActionContext,
	owner, repoName string,
	results *scanner.ScanResults,
	spamUsers map[string]spamUserInfo,
	flags *ActionFlags,
) error {
	if len(results.Spam) == 0 || (!flags.autoClose && !flags.autoBlock) {
		return nil
	}

	fmt.Println("\n=== AUTOMATED ACTIONS ===")

	// Confirm with user
	if !confirmAction(len(results.Spam), len(spamUsers), flags.autoClose, flags.autoBlock, flags.githubBlock) {
		fmt.Println("Actions cancelled by user.")
		return nil
	}

	// Block users first
	if flags.autoBlock {
		executeBlockActions(ctx, spamUsers, flags.githubBlock)
	}

	// Close PRs
	if flags.autoClose {
		executeCloseActions(ctx, owner, repoName, results)
	}

	fmt.Println("\n✓ Automated actions completed")
	return nil
}

// displayActionSuggestions shows suggestions if no automated actions were taken
func displayActionSuggestions(repo string, hasSpam, autoClose, autoBlock, githubBlock bool) {
	if !hasSpam || autoClose || autoBlock {
		return
	}

	fmt.Println("\nTo take action automatically, use:")
	fmt.Printf("  prguard scan %s --auto-close --auto-block\n", repo)
	if githubBlock {
		fmt.Printf("  Add --github-block to also block on GitHub\n")
	}
}
