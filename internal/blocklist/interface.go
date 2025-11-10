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

package blocklist

import "github.com/prguard/prguard/pkg/models"

// BlocklistManager defines the interface for managing the blocklist
type BlocklistManager interface {
	// Block operations
	Block(username, reason, evidenceURL, blockedBy, severity, source string) (*models.BlocklistEntry, error)
	Unblock(username string) error
	IsBlocked(username string) (bool, error)

	// Query operations
	List() ([]*models.BlocklistEntry, error)
	GetByUsername(username string) ([]*models.BlocklistEntry, error)

	// Import/Export operations
	ExportJSON(path string) error
	ExportCSV(path string) error
	ImportJSON(path string) (int, error)
	ImportJSONFromURL(url string) (int, error)
}
