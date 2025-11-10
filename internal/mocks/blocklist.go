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

import "github.com/prguard/prguard/pkg/models"

// MockBlocklistManager is a mock implementation of blocklist.BlocklistManager for testing
type MockBlocklistManager struct {
	BlockFn             func(username, reason, evidenceURL, blockedBy, severity, source string) (*models.BlocklistEntry, error)
	UnblockFn           func(username string) error
	IsBlockedFn         func(username string) (bool, error)
	ListFn              func() ([]*models.BlocklistEntry, error)
	GetByUsernameFn     func(username string) ([]*models.BlocklistEntry, error)
	ExportJSONFn        func(path string) error
	ExportCSVFn         func(path string) error
	ImportJSONFn        func(path string) (int, error)
	ImportJSONFromURLFn func(url string) (int, error)
}

func (m *MockBlocklistManager) Block(username, reason, evidenceURL, blockedBy, severity, source string) (*models.BlocklistEntry, error) {
	if m.BlockFn != nil {
		return m.BlockFn(username, reason, evidenceURL, blockedBy, severity, source)
	}
	return models.NewBlocklistEntry(username, reason, evidenceURL, blockedBy, severity, source), nil
}

func (m *MockBlocklistManager) Unblock(username string) error {
	if m.UnblockFn != nil {
		return m.UnblockFn(username)
	}
	return nil
}

func (m *MockBlocklistManager) IsBlocked(username string) (bool, error) {
	if m.IsBlockedFn != nil {
		return m.IsBlockedFn(username)
	}
	return false, nil
}

func (m *MockBlocklistManager) List() ([]*models.BlocklistEntry, error) {
	if m.ListFn != nil {
		return m.ListFn()
	}
	return []*models.BlocklistEntry{}, nil
}

func (m *MockBlocklistManager) GetByUsername(username string) ([]*models.BlocklistEntry, error) {
	if m.GetByUsernameFn != nil {
		return m.GetByUsernameFn(username)
	}
	return []*models.BlocklistEntry{}, nil
}

func (m *MockBlocklistManager) ExportJSON(path string) error {
	if m.ExportJSONFn != nil {
		return m.ExportJSONFn(path)
	}
	return nil
}

func (m *MockBlocklistManager) ExportCSV(path string) error {
	if m.ExportCSVFn != nil {
		return m.ExportCSVFn(path)
	}
	return nil
}

func (m *MockBlocklistManager) ImportJSON(path string) (int, error) {
	if m.ImportJSONFn != nil {
		return m.ImportJSONFn(path)
	}
	return 0, nil
}

func (m *MockBlocklistManager) ImportJSONFromURL(url string) (int, error) {
	if m.ImportJSONFromURLFn != nil {
		return m.ImportJSONFromURLFn(url)
	}
	return 0, nil
}
