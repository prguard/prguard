// Copyright 2025 Logan Lindquist Land
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scanner

import (
	"testing"
	"time"

	"github.com/prguard/prguard/internal/config"
	"github.com/prguard/prguard/internal/github"
)

func getTestConfig() *config.Config {
	cfg := &config.Config{
		Filters: config.FiltersConfig{
			MinFiles:        2,
			MinLines:        10,
			AccountAgeDays:  7,
			ReadmeOnlyBlock: true,
			Whitelist:       []string{"dependabot[bot]", "renovate[bot]"},
			SpamPhrases:     []string{"click here", "visit my site"},
		},
	}
	return cfg
}

func TestIsSingleFileReadmeEdit(t *testing.T) {
	scanner := NewScanner(getTestConfig())

	tests := []struct {
		name     string
		pr       *github.PullRequest
		expected bool
	}{
		{
			name: "Single README.md file",
			pr: &github.PullRequest{
				FilesCount: 1,
				Files:      []string{"README.md"},
			},
			expected: true,
		},
		{
			name: "Single readme.txt file (case insensitive)",
			pr: &github.PullRequest{
				FilesCount: 1,
				Files:      []string{"readme.txt"},
			},
			expected: true,
		},
		{
			name: "Multiple files including README",
			pr: &github.PullRequest{
				FilesCount: 2,
				Files:      []string{"README.md", "src/main.go"},
			},
			expected: false,
		},
		{
			name: "Single non-README file",
			pr: &github.PullRequest{
				FilesCount: 1,
				Files:      []string{"src/main.go"},
			},
			expected: false,
		},
		{
			name: "No files",
			pr: &github.PullRequest{
				FilesCount: 0,
				Files:      []string{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.isSingleFileReadmeEdit(tt.pr)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsNewAccount(t *testing.T) {
	scanner := NewScanner(getTestConfig())

	tests := []struct {
		name     string
		user     *github.User
		expected bool
	}{
		{
			name: "Account created 1 day ago",
			user: &github.User{
				Login:     "newuser",
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
			expected: true,
		},
		{
			name: "Account created 6 days ago",
			user: &github.User{
				Login:     "recentuser",
				CreatedAt: time.Now().Add(-6 * 24 * time.Hour),
			},
			expected: true,
		},
		{
			name: "Account created 8 days ago",
			user: &github.User{
				Login:     "olduser",
				CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Account created 1 year ago",
			user: &github.User{
				Login:     "establisheduser",
				CreatedAt: time.Now().Add(-365 * 24 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.isNewAccount(tt.user)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsMinimalChanges(t *testing.T) {
	scanner := NewScanner(getTestConfig())

	tests := []struct {
		name     string
		pr       *github.PullRequest
		expected bool
	}{
		{
			name: "Single file with few lines",
			pr: &github.PullRequest{
				FilesCount: 1,
				Additions:  3,
				Deletions:  2,
			},
			expected: true,
		},
		{
			name: "Multiple files but few lines",
			pr: &github.PullRequest{
				FilesCount: 2,
				Additions:  4,
				Deletions:  3,
			},
			expected: true,
		},
		{
			name: "Single file but many lines",
			pr: &github.PullRequest{
				FilesCount: 1,
				Additions:  50,
				Deletions:  10,
			},
			expected: true,
		},
		{
			name: "Multiple files and many lines",
			pr: &github.PullRequest{
				FilesCount: 5,
				Additions:  100,
				Deletions:  50,
			},
			expected: false,
		},
		{
			name: "Exactly at threshold (2 files, 10 lines)",
			pr: &github.PullRequest{
				FilesCount: 2,
				Additions:  6,
				Deletions:  4,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.isMinimalChanges(tt.pr)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for %d files, %d lines",
					tt.expected, result, tt.pr.FilesCount, tt.pr.Additions+tt.pr.Deletions)
			}
		})
	}
}

func TestContainsSpamPhrases(t *testing.T) {
	scanner := NewScanner(getTestConfig())

	tests := []struct {
		name     string
		pr       *github.PullRequest
		expected bool
	}{
		{
			name: "Title contains spam phrase",
			pr: &github.PullRequest{
				Title: "Please click here for more info",
				Body:  "Some description",
			},
			expected: true,
		},
		{
			name: "Body contains spam phrase",
			pr: &github.PullRequest{
				Title: "Update documentation",
				Body:  "Visit my site for details",
			},
			expected: true,
		},
		{
			name: "Case insensitive match",
			pr: &github.PullRequest{
				Title: "CLICK HERE NOW",
				Body:  "",
			},
			expected: true,
		},
		{
			name: "No spam phrases",
			pr: &github.PullRequest{
				Title: "Fix bug in authentication",
				Body:  "This PR fixes the login issue",
			},
			expected: false,
		},
		{
			name: "Empty PR",
			pr: &github.PullRequest{
				Title: "",
				Body:  "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.containsSpamPhrases(tt.pr)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsWhitelisted(t *testing.T) {
	scanner := NewScanner(getTestConfig())

	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{
			name:     "Dependabot is whitelisted",
			username: "dependabot[bot]",
			expected: true,
		},
		{
			name:     "Renovate is whitelisted",
			username: "renovate[bot]",
			expected: true,
		},
		{
			name:     "Regular user not whitelisted",
			username: "regular-user",
			expected: false,
		},
		{
			name:     "Empty username",
			username: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.isWhitelisted(tt.username)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestScanPR_DefiniteSpam(t *testing.T) {
	scanner := NewScanner(getTestConfig())

	tests := []struct {
		name         string
		pr           *github.PullRequest
		user         *github.User
		expectSpam   bool
		expectReason string
	}{
		{
			name: "Single README edit by new account",
			pr: &github.PullRequest{
				Number:     1,
				Title:      "Update README",
				Author:     "newspammer",
				FilesCount: 1,
				Files:      []string{"README.md"},
				Additions:  5,
			},
			user: &github.User{
				Login:     "newspammer",
				CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
			},
			expectSpam:   true,
			expectReason: "Single-file README-only edit",
		},
		{
			name: "README edit with spam phrase",
			pr: &github.PullRequest{
				Number:     2,
				Title:      "Update docs - click here",
				Author:     "spammer",
				FilesCount: 1,
				Files:      []string{"README.md"},
				Additions:  3,
			},
			user: &github.User{
				Login:     "spammer",
				CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
			},
			expectSpam:   true,
			expectReason: "Contains spam phrases",
		},
		{
			name: "Regular PR by established user",
			pr: &github.PullRequest{
				Number:     3,
				Title:      "Add new feature",
				Author:     "contributor",
				FilesCount: 5,
				Files:      []string{"src/main.go", "README.md", "test.go"},
				Additions:  100,
			},
			user: &github.User{
				Login:     "contributor",
				CreatedAt: time.Now().Add(-365 * 24 * time.Hour),
			},
			expectSpam: false,
		},
		{
			name: "Whitelisted bot",
			pr: &github.PullRequest{
				Number:     4,
				Title:      "Update dependencies",
				Author:     "dependabot[bot]",
				FilesCount: 1,
				Files:      []string{"go.mod"},
				Additions:  1,
			},
			user: &github.User{
				Login:     "dependabot[bot]",
				CreatedAt: time.Now().Add(-100 * 24 * time.Hour),
			},
			expectSpam: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.ScanPR(tt.pr, tt.user)

			if result.IsSpam != tt.expectSpam {
				t.Errorf("Expected IsSpam=%v, got %v. Reasons: %v",
					tt.expectSpam, result.IsSpam, result.Reasons)
			}

			if tt.expectSpam && tt.expectReason != "" {
				found := false
				for _, reason := range result.Reasons {
					if reason == tt.expectReason {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected reason '%s' not found in: %v", tt.expectReason, result.Reasons)
				}
			}
		})
	}
}

func TestScanPR_UncertainCases(t *testing.T) {
	scanner := NewScanner(getTestConfig())

	tests := []struct {
		name            string
		pr              *github.PullRequest
		user            *github.User
		expectUncertain bool
	}{
		{
			name: "New account with minimal changes (not README)",
			pr: &github.PullRequest{
				Number:     1,
				Title:      "Minor fix",
				Author:     "newuser",
				FilesCount: 1,
				Files:      []string{"src/util.go"},
				Additions:  3,
			},
			user: &github.User{
				Login:     "newuser",
				CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
			},
			expectUncertain: true,
		},
		{
			name: "Established user with minimal changes",
			pr: &github.PullRequest{
				Number:     2,
				Title:      "Typo fix",
				Author:     "olduser",
				FilesCount: 1,
				Files:      []string{"docs/guide.md"},
				Additions:  1,
			},
			user: &github.User{
				Login:     "olduser",
				CreatedAt: time.Now().Add(-365 * 24 * time.Hour),
			},
			expectUncertain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.ScanPR(tt.pr, tt.user)

			if tt.expectUncertain && !result.IsUncertain && !result.IsSpam {
				t.Errorf("Expected uncertain or spam, got clean. Reasons: %v", result.Reasons)
			}

			if result.IsUncertain && result.RecommendAction != "Manual review recommended" {
				t.Errorf("Uncertain PR should recommend manual review, got: %s", result.RecommendAction)
			}
		})
	}
}
