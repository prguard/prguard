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
	"testing"
)

func TestParseRepo(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "valid repo",
			input:     "owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "valid repo with hyphens",
			input:     "my-org/my-repo",
			wantOwner: "my-org",
			wantName:  "my-repo",
			wantErr:   false,
		},
		{
			name:      "valid repo with underscores",
			input:     "my_org/my_repo",
			wantOwner: "my_org",
			wantName:  "my_repo",
			wantErr:   false,
		},
		{
			name:      "valid repo with dots",
			input:     "org.name/repo.name",
			wantOwner: "org.name",
			wantName:  "repo.name",
			wantErr:   false,
		},
		{
			name:    "missing slash",
			input:   "ownerrepo",
			wantErr: true,
		},
		{
			name:      "empty owner",
			input:     "/repo",
			wantOwner: "",
			wantName:  "repo",
			wantErr:   false, // parseRepo allows this, validation happens elsewhere
		},
		{
			name:      "empty repo",
			input:     "owner/",
			wantOwner: "owner",
			wantName:  "",
			wantErr:   false, // parseRepo allows this, validation happens elsewhere
		},
		{
			name:      "multiple slashes",
			input:     "owner/repo/extra",
			wantOwner: "owner",
			wantName:  "repo/extra", // Takes first slash
			wantErr:   false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := parseRepo(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}

			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}

func TestScanCommand_Flags(t *testing.T) {
	configPath := "config.yaml"
	cmd := NewScanCommand(&configPath)

	if cmd.Use != "scan <owner>/<repo>" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	// Test that command requires exactly 1 arg
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with no args")
	}

	// Test with 2 args (should fail)
	cmd.SetArgs([]string{"owner/repo", "extra"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with too many args")
	}
}

func TestScanAllCommand_Flags(t *testing.T) {
	configPath := "config.yaml"
	cmd := NewScanAllCommand(&configPath)

	if cmd.Use != "scan-all" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	// scan-all takes no args
	cmd.SetArgs([]string{"extra"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error with args")
	}
}

// NOTE: Full integration tests for scan commands would require:
// - Mocking the GitHub client to return fake PR/user data
// - Mocking the scanner to return fake scan results
// - This would require refactoring to use dependency injection with interfaces
//
// Current architecture directly instantiates:
//   - github.NewClient() in initClients()
//   - scanner.NewScanner() in runScan()
//
// For full test coverage, consider:
// 1. Creating interfaces: GitHubClient, Scanner
// 2. Passing implementations via dependency injection
// 3. Using test doubles in tests
