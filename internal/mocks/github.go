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

package mocks

import "github.com/prguard/prguard/internal/github"

// MockGitHubClient is a mock implementation of github.GitHubClient for testing
type MockGitHubClient struct {
	GetPullRequestsFn   func(owner, repo string) ([]*github.PullRequest, error)
	ClosePullRequestFn  func(owner, repo string, number int, comment string) error
	AddLabelFn          func(owner, repo string, number int, label string) error
	GetUserFn           func(username string) (*github.User, error)
	BlockUserOrgFn      func(org, username string) error
	BlockUserPersonalFn func(username string) error
}

func (m *MockGitHubClient) GetPullRequests(owner, repo string) ([]*github.PullRequest, error) {
	if m.GetPullRequestsFn != nil {
		return m.GetPullRequestsFn(owner, repo)
	}
	return nil, nil
}

func (m *MockGitHubClient) ClosePullRequest(owner, repo string, number int, comment string) error {
	if m.ClosePullRequestFn != nil {
		return m.ClosePullRequestFn(owner, repo, number, comment)
	}
	return nil
}

func (m *MockGitHubClient) AddLabel(owner, repo string, number int, label string) error {
	if m.AddLabelFn != nil {
		return m.AddLabelFn(owner, repo, number, label)
	}
	return nil
}

func (m *MockGitHubClient) GetUser(username string) (*github.User, error) {
	if m.GetUserFn != nil {
		return m.GetUserFn(username)
	}
	return nil, nil
}

func (m *MockGitHubClient) BlockUserOrg(org, username string) error {
	if m.BlockUserOrgFn != nil {
		return m.BlockUserOrgFn(org, username)
	}
	return nil
}

func (m *MockGitHubClient) BlockUserPersonal(username string) error {
	if m.BlockUserPersonalFn != nil {
		return m.BlockUserPersonalFn(username)
	}
	return nil
}
