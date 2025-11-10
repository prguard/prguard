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
		RunE: func(cmd *cobra.Command, args []string) error {
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
	defer db.Close()

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
