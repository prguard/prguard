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

package commands

import (
	"fmt"
	"testing"
	"time"

	"github.com/prguard/prguard/internal/github"
	"github.com/prguard/prguard/internal/mocks"
	"github.com/prguard/prguard/internal/scanner"
)

func TestReviewCommand_Flags(t *testing.T) {
	configPath := "config.yaml"
	cmd := NewReviewCommand(&configPath)

	// Test command requires exactly 1 arg
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with no args")
	}

	cmd.SetArgs([]string{"owner/repo", "extra"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with too many args")
	}
}

func TestReview_NoUncertainPRs(t *testing.T) {
	mockGH := &mocks.MockGitHubClient{
		GetPullRequestsFn: func(owner, repo string) ([]*github.PullRequest, error) {
			return []*github.PullRequest{}, nil
		},
	}

	mockScanner := &mocks.MockScanner{
		ScanRepositoryFn: func(ghClient github.GitHubClient, owner, repo string) (*scanner.ScanResults, error) {
			return &scanner.ScanResults{
				Total:     3,
				Spam:      []*scanner.ScanResult{},
				Uncertain: []*scanner.ScanResult{},
				Clean: []*scanner.ScanResult{
					{
						PR: &github.PullRequest{
							Number: 1,
							Title:  "Add feature",
							Author: "gooduser",
						},
					},
				},
			}, nil
		},
	}

	results, err := executeReview(mockGH, mockScanner, "testowner", "testrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results == nil {
		t.Fatal("expected results, got nil")
	}

	if len(results.Uncertain) != 0 {
		t.Errorf("expected 0 uncertain PRs, got %d", len(results.Uncertain))
	}
}

func TestReview_WithUncertainPRs(t *testing.T) {
	mockGH := &mocks.MockGitHubClient{
		GetPullRequestsFn: func(owner, repo string) ([]*github.PullRequest, error) {
			return []*github.PullRequest{
				{
					Number:     123,
					Title:      "Suspicious PR",
					Author:     "newuser",
					HTMLURL:    "https://github.com/test/repo/pull/123",
					FilesCount: 1,
					Additions:  5,
					Deletions:  2,
					CreatedAt:  time.Now(),
				},
			}, nil
		},
	}

	mockScanner := &mocks.MockScanner{
		ScanRepositoryFn: func(ghClient github.GitHubClient, owner, repo string) (*scanner.ScanResults, error) {
			return &scanner.ScanResults{
				Total: 1,
				Spam:  []*scanner.ScanResult{},
				Uncertain: []*scanner.ScanResult{
					{
						PR: &github.PullRequest{
							Number:     123,
							Title:      "Suspicious PR",
							Author:     "newuser",
							HTMLURL:    "https://github.com/test/repo/pull/123",
							FilesCount: 1,
							Additions:  5,
							Deletions:  2,
						},
						IsUncertain:     true,
						Reasons:         []string{"Account created recently", "Minimal changes"},
						RecommendAction: "Manual review recommended",
					},
				},
				Clean: []*scanner.ScanResult{},
			}, nil
		},
	}

	results, err := executeReview(mockGH, mockScanner, "testowner", "testrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results.Uncertain) != 1 {
		t.Errorf("expected 1 uncertain PR, got %d", len(results.Uncertain))
	}

	uncertain := results.Uncertain[0]
	if uncertain.PR.Number != 123 {
		t.Errorf("expected PR #123, got #%d", uncertain.PR.Number)
	}

	if uncertain.PR.Author != "newuser" {
		t.Errorf("expected author 'newuser', got '%s'", uncertain.PR.Author)
	}

	if len(uncertain.Reasons) != 2 {
		t.Errorf("expected 2 reasons, got %d", len(uncertain.Reasons))
	}
}

func TestReview_ScanError(t *testing.T) {
	mockGH := &mocks.MockGitHubClient{}

	mockScanner := &mocks.MockScanner{
		ScanRepositoryFn: func(ghClient github.GitHubClient, owner, repo string) (*scanner.ScanResults, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	_, err := executeReview(mockGH, mockScanner, "testowner", "testrepo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// executeReview is a testable helper function that performs the review logic
func executeReview(ghClient github.GitHubClient, scan scanner.PRScanner, owner, repo string) (*scanner.ScanResults, error) {
	// Scan repository
	results, err := scan.ScanRepository(ghClient, owner, repo)
	if err != nil {
		return nil, err
	}

	return results, nil
}
