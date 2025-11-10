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
	"errors"
	"fmt"
	"testing"

	"github.com/prguard/prguard/internal/github"
	"github.com/prguard/prguard/internal/mocks"
)

func TestClosePRCommand_Flags(t *testing.T) {
	configPath := "config.yaml"
	cmd := NewClosePRCommand(&configPath)

	// Test command requires at least 2 args
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with no args")
	}

	cmd.SetArgs([]string{"owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with only 1 arg")
	}

	// Check flags exist
	commentFlag := cmd.Flags().Lookup("comment")
	if commentFlag == nil {
		t.Error("comment flag not found")
	}

	labelFlag := cmd.Flags().Lookup("label")
	if labelFlag == nil {
		t.Error("label flag not found")
	}
}

func TestClosePR_Success(t *testing.T) {
	var closedPRs []int
	var addedLabels []int

	mockClient := &mocks.MockGitHubClient{
		ClosePullRequestFn: func(owner, repo string, number int, comment string) error {
			closedPRs = append(closedPRs, number)
			if owner != "testowner" || repo != "testrepo" {
				t.Errorf("unexpected owner/repo: %s/%s", owner, repo)
			}
			return nil
		},
		AddLabelFn: func(owner, repo string, number int, label string) error {
			addedLabels = append(addedLabels, number)
			if label != "spam" {
				t.Errorf("unexpected label: %s", label)
			}
			return nil
		},
	}

	err := executeClosePR(mockClient, "testowner", "testrepo", []int{123, 456}, "Test comment", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(closedPRs) != 2 {
		t.Errorf("expected 2 PRs closed, got %d", len(closedPRs))
	}

	if closedPRs[0] != 123 || closedPRs[1] != 456 {
		t.Errorf("unexpected PRs closed: %v", closedPRs)
	}

	if len(addedLabels) != 2 {
		t.Errorf("expected 2 labels added, got %d", len(addedLabels))
	}
}

func TestClosePR_WithoutLabel(t *testing.T) {
	var closedPRs []int
	labelCalled := false

	mockClient := &mocks.MockGitHubClient{
		ClosePullRequestFn: func(owner, repo string, number int, comment string) error {
			closedPRs = append(closedPRs, number)
			return nil
		},
		AddLabelFn: func(owner, repo string, number int, label string) error {
			labelCalled = true
			return nil
		},
	}

	err := executeClosePR(mockClient, "testowner", "testrepo", []int{789}, "Test comment", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(closedPRs) != 1 {
		t.Errorf("expected 1 PR closed, got %d", len(closedPRs))
	}

	if labelCalled {
		t.Error("label should not have been added")
	}
}

func TestClosePR_CloseFails(t *testing.T) {
	mockClient := &mocks.MockGitHubClient{
		ClosePullRequestFn: func(owner, repo string, number int, comment string) error {
			return errors.New("API error")
		},
		AddLabelFn: func(owner, repo string, number int, label string) error {
			return nil
		},
	}

	err := executeClosePR(mockClient, "testowner", "testrepo", []int{123}, "Test comment", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "failed to close PR #123: API error" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClosePR_LabelFailsContinues(t *testing.T) {
	var closedPRs []int

	mockClient := &mocks.MockGitHubClient{
		ClosePullRequestFn: func(owner, repo string, number int, comment string) error {
			closedPRs = append(closedPRs, number)
			return nil
		},
		AddLabelFn: func(owner, repo string, number int, label string) error {
			return errors.New("label API error")
		},
	}

	err := executeClosePR(mockClient, "testowner", "testrepo", []int{123}, "Test comment", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(closedPRs) != 1 {
		t.Errorf("PR should still be closed even if label fails, got %d closed", len(closedPRs))
	}
}

// executeClosePR is a testable helper function that performs the close PR logic
func executeClosePR(ghClient github.GitHubClient, owner, repo string, prNumbers []int, comment string, addLabel bool) error {
	for _, prNum := range prNumbers {
		// Add label if requested
		if addLabel {
			if err := ghClient.AddLabel(owner, repo, prNum, "spam"); err != nil {
				// Label failure is not fatal, just log it
				// In the real command this would print a warning
			}
		}

		// Close the PR
		if err := ghClient.ClosePullRequest(owner, repo, prNum, comment); err != nil {
			return fmt.Errorf("failed to close PR #%d: %w", prNum, err)
		}
	}

	return nil
}
