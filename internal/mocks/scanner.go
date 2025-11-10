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
