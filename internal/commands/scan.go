package commands

import (
	"fmt"

	"github.com/prguard/prguard/internal/scanner"
	"github.com/spf13/cobra"
)

// NewScanCommand creates the scan command
func NewScanCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan <owner>/<repo>",
		Short: "Scan a repository for spam pull requests",
		Long:  `Analyzes all open pull requests in a repository for spam indicators`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(*configPath, args[0])
		},
	}
	return cmd
}

func runScan(configPath, repo string) error {
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

	fmt.Printf("Scanning repository %s/%s...\n\n", owner, repoName)

	// Create scanner
	scan := scanner.NewScanner(cfg)

	// Scan repository
	results, err := scan.ScanRepository(ghClient, owner, repoName)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Display results
	fmt.Printf("Total PRs: %d\n", results.Total)
	fmt.Printf("Spam detected: %d\n", len(results.Spam))
	fmt.Printf("Uncertain: %d\n", len(results.Uncertain))
	fmt.Printf("Clean: %d\n\n", len(results.Clean))

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

	return nil
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
