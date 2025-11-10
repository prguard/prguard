package scanner

import "github.com/prguard/prguard/internal/github"

// PRScanner defines the interface for scanning pull requests for spam
type PRScanner interface {
	ScanRepository(ghClient github.GitHubClient, owner, repo string) (*ScanResults, error)
}
