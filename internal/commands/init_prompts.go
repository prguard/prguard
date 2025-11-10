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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// determineConfigPath determines where the config file should be created
func determineConfigPath(global bool) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir := filepath.Join(home, ".config", "prguard")
		return filepath.Join(configDir, "config.yaml"), nil
	}
	return "./config.yaml", nil
}

// promptOverwriteExisting checks if config exists and prompts for overwrite
func promptOverwriteExisting(path string, reader *bufio.Reader) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		// File doesn't exist, proceed
		return true, nil
	}

	fmt.Printf("Config file already exists at: %s\n", path)
	fmt.Print("Overwrite? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Initialization cancelled.")
		return false, nil
	}

	return true, nil
}

// promptGitHubToken prompts the user for their GitHub token
func promptGitHubToken(reader *bufio.Reader) (string, error) {
	fmt.Println("GitHub Personal Access Token:")
	fmt.Println("  Create one at: https://github.com/settings/tokens")
	fmt.Println("  Required scopes: repo, write:discussion")
	fmt.Print("Token: ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	if token == "" {
		return "", fmt.Errorf("GitHub token is required")
	}

	return token, nil
}

// promptOrgOrUser prompts for GitHub organization or user
func promptOrgOrUser(reader *bufio.Reader, gitUser string) (org, user string, err error) {
	fmt.Println("\nGitHub Organization or User:")
	if gitUser != "" {
		fmt.Printf("  Detected from git config: %s\n", gitUser)
	}
	fmt.Print("Enter org name (or press Enter for user mode): ")
	org, _ = reader.ReadString('\n')
	org = strings.TrimSpace(org)

	// If no org, we need a username
	if org == "" {
		user, err = promptUsername(reader, gitUser)
		if err != nil {
			return "", "", err
		}
	}

	return org, user, nil
}

// promptUsername prompts for GitHub username with optional default
func promptUsername(reader *bufio.Reader, gitUser string) (string, error) {
	var user string

	if gitUser != "" {
		fmt.Printf("Use '%s' as username? (Y/n): ", gitUser)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "" || response == "y" || response == "yes" {
			user = gitUser
		}
	}

	if user == "" {
		fmt.Print("GitHub username: ")
		user, _ = reader.ReadString('\n')
		user = strings.TrimSpace(user)
	}

	if user == "" {
		return "", fmt.Errorf("either org or user is required")
	}

	return user, nil
}

// promptRepositories prompts for repositories to monitor
func promptRepositories(reader *bufio.Reader) ([]string, error) {
	fmt.Println("\nRepositories to monitor (optional):")
	fmt.Println("  Enter repositories in 'owner/repo' format, one per line")
	fmt.Println("  Press Enter on empty line to finish")

	var repos []string
	for {
		fmt.Print("Repository (or Enter to skip): ")
		repo, _ := reader.ReadString('\n')
		repo = strings.TrimSpace(repo)

		if repo == "" {
			break
		}

		// Validate format
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			fmt.Println("  Invalid format. Use: owner/repo")
			continue
		}

		repos = append(repos, repo)
	}

	return repos, nil
}

// writeConfigFile writes the config to disk and displays success message
func writeConfigFile(path, content string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// displayInitSuccess shows success message and next steps
func displayInitSuccess(configPath string) {
	fmt.Printf("\n✓ Configuration created at: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review and edit the config file if needed")
	fmt.Println("  2. Run: prguard scan owner/repo")
	fmt.Println("  3. Or: prguard scan-all (if you added repositories)")
}
