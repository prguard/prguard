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

package github

// GitHubClient defines the interface for GitHub operations needed by commands
type GitHubClient interface {
	// PR operations
	GetPullRequests(owner, repo string) ([]*PullRequest, error)
	ClosePullRequest(owner, repo string, number int, comment string) error
	AddLabel(owner, repo string, number int, label string) error

	// User operations
	GetUser(username string) (*User, error)
	BlockUserOrg(org, username string) error
	BlockUserPersonal(username string) error
}
