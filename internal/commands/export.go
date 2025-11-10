package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewExportCommand creates the export command
func NewExportCommand(configPath *string) *cobra.Command {
	var format, output string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the blocklist to a file",
		Long:  `Exports the blocklist to JSON or CSV format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(*configPath, format, output)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "json", "Export format (json or csv)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: blocklist.json or blocklist.csv)")

	return cmd
}

func runExport(configPath, format, output string) error {
	_, _, blManager, db, err := initClients(configPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Determine output path
	if output == "" {
		if format == "json" {
			output = "blocklist.json"
		} else if format == "csv" {
			output = "blocklist.csv"
		} else {
			return fmt.Errorf("invalid format, must be json or csv")
		}
	}

	// Export
	switch format {
	case "json":
		if err := blManager.ExportJSON(output); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
	case "csv":
		if err := blManager.ExportCSV(output); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
	default:
		return fmt.Errorf("invalid format, must be json or csv")
	}

	absPath, _ := filepath.Abs(output)
	fmt.Printf("✓ Blocklist exported to %s\n", absPath)

	return nil
}
