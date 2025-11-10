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

	"github.com/spf13/cobra"
)

// NewImportCommand creates the import command
func NewImportCommand(configPath *string) *cobra.Command {
	var file, url string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import blocklist entries from a file or URL",
		Long:  `Imports blocklist entries from a JSON file or remote URL`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runImport(*configPath, file, url)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file to import")
	cmd.Flags().StringVarP(&url, "url", "u", "", "URL to JSON file to import")

	return cmd
}

func runImport(configPath, file, url string) error {
	if file == "" && url == "" {
		return fmt.Errorf("either --file or --url must be specified")
	}
	if file != "" && url != "" {
		return fmt.Errorf("cannot specify both --file and --url")
	}

	_, _, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck

	var imported int

	if file != "" {
		fmt.Printf("Importing from file: %s\n", file)
		imported, err = blManager.ImportJSON(file)
	} else {
		fmt.Printf("Importing from URL: %s\n", url)
		imported, err = blManager.ImportJSONFromURL(url)
	}

	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	fmt.Printf("✓ Successfully imported %d %s\n", imported, pluralize("entry", "entries", imported))

	return nil
}

func pluralize(singular, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}
