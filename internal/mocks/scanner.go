package mocks

import (
	"github.com/prguard/prguard/internal/github"
	"github.com/prguard/prguard/internal/scanner"
)

// MockScanner is a mock implementation of scanner.PRScanner for testing
type MockScanner struct {
	ScanRepositoryFn func(ghClient github.GitHubClient, owner, repo string) (*scanner.ScanResults, error)
}

func (m *MockScanner) ScanRepository(ghClient github.GitHubClient, owner, repo string) (*scanner.ScanResults, error) {
	if m.ScanRepositoryFn != nil {
		return m.ScanRepositoryFn(ghClient, owner, repo)
	}
	return &scanner.ScanResults{
		Total:     0,
		Spam:      []*scanner.ScanResult{},
		Uncertain: []*scanner.ScanResult{},
		Clean:     []*scanner.ScanResult{},
	}, nil
}
