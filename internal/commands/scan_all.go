package commands

import (
	"fmt"

	"github.com/prguard/prguard/internal/scanner"
	"github.com/spf13/cobra"
)

// NewScanAllCommand creates the scan-all command
func NewScanAllCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan-all",
		Short: "Scan all configured repositories for spam pull requests",
		Long:  `Scans all repositories listed in the configuration file for spam indicators`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScanAll(*configPath)
		},
	}
	return cmd
}

func runScanAll(configPath string) error {
	cfg, ghClient, _, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Add repositories to your config.yaml file")
	}

	fmt.Printf("Scanning %d configured repositories...\n\n", len(cfg.Repositories))

	// Create scanner
	scan := scanner.NewScanner(cfg)

	totalRepos := 0
	totalPRs := 0
	totalSpam := 0
	totalUncertain := 0

	for _, repo := range cfg.Repositories {
		fmt.Printf("=== %s ===\n", repo.FullName())

		// Scan repository
		results, err := scan.ScanRepository(ghClient, repo.Owner, repo.Name)
		if err != nil {
			fmt.Printf("Error scanning %s: %v\n\n", repo.FullName(), err)
			continue
		}

		totalRepos++
		totalPRs += results.Total
		totalSpam += len(results.Spam)
		totalUncertain += len(results.Uncertain)

		fmt.Printf("Total PRs: %d | Spam: %d | Uncertain: %d | Clean: %d\n",
			results.Total, len(results.Spam), len(results.Uncertain), len(results.Clean))

		if len(results.Spam) > 0 {
			fmt.Println("\nSpam detected:")
			for _, result := range results.Spam {
				fmt.Printf("  - PR #%d: %s (by %s)\n", result.PR.Number, result.PR.Title, result.PR.Author)
			}
		}

		if len(results.Uncertain) > 0 {
			fmt.Println("\nNeeds review:")
			for _, result := range results.Uncertain {
				fmt.Printf("  - PR #%d: %s (by %s)\n", result.PR.Number, result.PR.Title, result.PR.Author)
			}
		}

		fmt.Println()
	}

	fmt.Println("=== SUMMARY ===")
	fmt.Printf("Repositories scanned: %d\n", totalRepos)
	fmt.Printf("Total PRs: %d\n", totalPRs)
	fmt.Printf("Spam detected: %d\n", totalSpam)
	fmt.Printf("Needs review: %d\n", totalUncertain)

	if totalSpam > 0 || totalUncertain > 0 {
		fmt.Println("\nUse 'prguard scan <owner>/<repo>' for detailed results")
		fmt.Println("Use 'prguard block <username>' to block a user")
	}

	return nil
}
