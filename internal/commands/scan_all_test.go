package commands

import (
	"fmt"
	"testing"
)

func TestScanAllCommand_FlagExistence(t *testing.T) {
	configPath := "config.yaml"
	cmd := NewScanAllCommand(&configPath)

	// Check flags exist
	autoCloseFlag := cmd.Flags().Lookup("auto-close")
	if autoCloseFlag == nil {
		t.Error("auto-close flag not found")
	}

	autoBlockFlag := cmd.Flags().Lookup("auto-block")
	if autoBlockFlag == nil {
		t.Error("auto-block flag not found")
	}

	githubBlockFlag := cmd.Flags().Lookup("github-block")
	if githubBlockFlag == nil {
		t.Error("github-block flag not found")
	}
}

func TestScanAll_FlagValidation(t *testing.T) {
	tests := []struct {
		name         string
		autoClose    bool
		autoBlock    bool
		githubBlock  bool
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid: no flags",
			autoClose:   false,
			autoBlock:   false,
			githubBlock: false,
			expectError: false,
		},
		{
			name:        "valid: auto-close only",
			autoClose:   true,
			autoBlock:   false,
			githubBlock: false,
			expectError: false,
		},
		{
			name:        "valid: auto-block only",
			autoClose:   false,
			autoBlock:   true,
			githubBlock: false,
			expectError: false,
		},
		{
			name:        "valid: auto-block with github-block",
			autoClose:   false,
			autoBlock:   true,
			githubBlock: true,
			expectError: false,
		},
		{
			name:         "invalid: github-block without auto-block",
			autoClose:    false,
			autoBlock:    false,
			githubBlock:  true,
			expectError:  true,
			errorMessage: "--github-block requires --auto-block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScanAllFlags(tt.autoClose, tt.autoBlock, tt.githubBlock)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if err.Error() != tt.errorMessage {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// validateScanAllFlags is extracted for testing
func validateScanAllFlags(autoClose, autoBlock, githubBlock bool) error {
	if githubBlock && !autoBlock {
		return fmt.Errorf("--github-block requires --auto-block")
	}
	return nil
}
