package scanner

import (
	"strings"
	"time"

	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/github"
)

// ScanResult represents the result of scanning a PR
type ScanResult struct {
	PR              *github.PullRequest
	IsSpam          bool
	IsUncertain     bool
	Reasons         []string
	Severity        string
	RecommendAction string
}

// Scanner analyzes pull requests for spam indicators
type Scanner struct {
	config *config.Config
}

// NewScanner creates a new PR scanner
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{config: cfg}
}

// ScanPR analyzes a pull request for spam indicators
func (s *Scanner) ScanPR(pr *github.PullRequest, user *github.User) *ScanResult {
	result := &ScanResult{
		PR:       pr,
		IsSpam:   false,
		Reasons:  []string{},
		Severity: "low",
	}

	// Check if user is whitelisted
	if s.isWhitelisted(pr.Author) {
		return result
	}

	// Check for single-file README edits
	if s.isSingleFileReadmeEdit(pr) {
		result.IsSpam = true
		result.Reasons = append(result.Reasons, "Single-file README-only edit")
		result.Severity = "high"
	}

	// Check account age
	if user != nil && s.isNewAccount(user) {
		if result.IsSpam {
			result.Reasons = append(result.Reasons, "Account created recently")
		} else {
			result.IsUncertain = true
			result.Reasons = append(result.Reasons, "Account created recently (suspicious but not definitive)")
		}
	}

	// Check for minimal changes
	if s.isMinimalChanges(pr) {
		if result.IsSpam {
			result.Reasons = append(result.Reasons, "Minimal changes (below threshold)")
		} else {
			result.IsUncertain = true
			result.Reasons = append(result.Reasons, "Minimal changes (below threshold)")
		}
	}

	// Check for spam phrases
	if s.containsSpamPhrases(pr) {
		result.IsSpam = true
		result.Reasons = append(result.Reasons, "Contains spam phrases")
		result.Severity = "high"
	}

	// Determine recommended action
	if result.IsSpam {
		result.RecommendAction = "Block user and close PR"
	} else if result.IsUncertain {
		result.RecommendAction = "Manual review recommended"
	} else {
		result.RecommendAction = "No action needed"
	}

	return result
}

// isWhitelisted checks if a user is in the whitelist
func (s *Scanner) isWhitelisted(username string) bool {
	for _, whitelisted := range s.config.Filters.Whitelist {
		if whitelisted == username {
			return true
		}
	}
	return false
}

// isSingleFileReadmeEdit checks if PR only modifies a single README file
func (s *Scanner) isSingleFileReadmeEdit(pr *github.PullRequest) bool {
	if !s.config.Filters.ReadmeOnlyBlock {
		return false
	}

	// Must be exactly one file
	if pr.FilesCount != 1 {
		return false
	}

	// Check if that file is a README
	for _, file := range pr.Files {
		if config.IsReadmeFile(file) {
			return true
		}
	}

	return false
}

// isNewAccount checks if the account was created recently
func (s *Scanner) isNewAccount(user *github.User) bool {
	threshold := time.Duration(s.config.Filters.AccountAgeDays) * 24 * time.Hour
	accountAge := time.Since(user.CreatedAt)
	return accountAge < threshold
}

// isMinimalChanges checks if the PR has minimal changes
func (s *Scanner) isMinimalChanges(pr *github.PullRequest) bool {
	totalLines := pr.Additions + pr.Deletions
	return pr.FilesCount < s.config.Filters.MinFiles || totalLines < s.config.Filters.MinLines
}

// containsSpamPhrases checks if PR title or body contains spam phrases
func (s *Scanner) containsSpamPhrases(pr *github.PullRequest) bool {
	if len(s.config.Filters.SpamPhrases) == 0 {
		return false
	}

	text := strings.ToLower(pr.Title + " " + pr.Body)
	for _, phrase := range s.config.Filters.SpamPhrases {
		if strings.Contains(text, strings.ToLower(phrase)) {
			return true
		}
	}

	return false
}

// ScanResults holds multiple scan results
type ScanResults struct {
	Total     int
	Spam      []*ScanResult
	Uncertain []*ScanResult
	Clean     []*ScanResult
}

// ScanRepository scans all open PRs in a repository
func (s *Scanner) ScanRepository(ghClient github.GitHubClient, owner, repo string) (*ScanResults, error) {
	prs, err := ghClient.GetPullRequests(owner, repo)
	if err != nil {
		return nil, err
	}

	results := &ScanResults{
		Total:     len(prs),
		Spam:      []*ScanResult{},
		Uncertain: []*ScanResult{},
		Clean:     []*ScanResult{},
	}

	for _, pr := range prs {
		// Fetch user information
		user, err := ghClient.GetUser(pr.Author)
		if err != nil {
			// If we can't fetch user info, continue with nil
			user = nil
		}

		scanResult := s.ScanPR(pr, user)

		if scanResult.IsSpam {
			results.Spam = append(results.Spam, scanResult)
		} else if scanResult.IsUncertain {
			results.Uncertain = append(results.Uncertain, scanResult)
		} else {
			results.Clean = append(results.Clean, scanResult)
		}
	}

	return results, nil
}
