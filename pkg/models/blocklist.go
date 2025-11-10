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

// Package models defines the data structures used throughout PRGuard.
package models

import (
	"time"

	"github.com/google/uuid"
)

// BlocklistEntry represents a blocked user in the database
type BlocklistEntry struct {
	ID          string    `json:"id" db:"id"`                     // UUID
	Username    string    `json:"username" db:"username"`         // GitHub username
	Reason      string    `json:"reason" db:"reason"`             // Reason for blocking
	EvidenceURL string    `json:"evidence_url" db:"evidence_url"` // Link to problematic PR/issue
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`       // When entry was created
	BlockedBy   string    `json:"blocked_by" db:"blocked_by"`     // Maintainer who added entry
	Severity    string    `json:"severity" db:"severity"`         // low/medium/high
	Source      string    `json:"source" db:"source"`             // manual/imported/auto-detected
	Metadata    string    `json:"metadata" db:"metadata"`         // JSON field for extensibility
}

// NewBlocklistEntry creates a new blocklist entry with a generated UUID
func NewBlocklistEntry(username, reason, evidenceURL, blockedBy, severity, source string) *BlocklistEntry {
	return &BlocklistEntry{
		ID:          uuid.New().String(),
		Username:    username,
		Reason:      reason,
		EvidenceURL: evidenceURL,
		Timestamp:   time.Now(),
		BlockedBy:   blockedBy,
		Severity:    severity,
		Source:      source,
		Metadata:    "{}",
	}
}

// Severity constants
const (
	SeverityLow    = "low"
	SeverityMedium = "medium"
	SeverityHigh   = "high"
)

// Source constants
const (
	SourceManual       = "manual"
	SourceImported     = "imported"
	SourceAutoDetected = "auto-detected"
)
