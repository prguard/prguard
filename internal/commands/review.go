package commands

import (
	"fmt"

	"github.com/prguard/prguard/internal/scanner"
	"github.com/spf13/cobra"
)

// NewReviewCommand creates the review command
func NewReviewCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <owner>/<repo>",
		Short: "Show PRs that need manual review",
		Long:  `Displays pull requests that have suspicious indicators but are not definitively spam`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReview(*configPath, args[0])
		},
	}
	return cmd
}

func runReview(configPath, repo string) error {
	cfg, ghClient, _, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Parse owner/repo
	owner, repoName, err := parseRepo(repo)
	if err != nil {
		return err
	}

	fmt.Printf("Scanning repository %s/%s for PRs needing review...\n\n", owner, repoName)

	// Create scanner
	scan := scanner.NewScanner(cfg)

	// Scan repository
	results, err := scan.ScanRepository(ghClient, owner, repoName)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(results.Uncertain) == 0 {
		fmt.Println("No PRs need manual review - all clear!")
		return nil
	}

	fmt.Printf("Found %d PR(s) needing manual review:\n\n", len(results.Uncertain))

	for i, result := range results.Uncertain {
		fmt.Printf("%d. PR #%d: %s\n", i+1, result.PR.Number, result.PR.Title)
		fmt.Printf("   Author: %s\n", result.PR.Author)
		fmt.Printf("   URL: %s\n", result.PR.HTMLURL)
		fmt.Printf("   Files changed: %d\n", result.PR.FilesCount)
		fmt.Printf("   Lines: +%d -%d\n", result.PR.Additions, result.PR.Deletions)
		fmt.Printf("   Suspicious indicators:\n")
		for _, reason := range result.Reasons {
			fmt.Printf("     - %s\n", reason)
		}
		fmt.Printf("   Recommendation: %s\n", result.RecommendAction)
		fmt.Println()
	}

	fmt.Println("To block a user: prguard block <username> --reason \"...\" --evidence <url>")
	fmt.Println("To close a PR: prguard close-pr <owner>/<repo> <pr-number>")

	return nil
}
