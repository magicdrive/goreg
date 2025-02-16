package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/magicdrive/goreg/internal/core"
)

func TestGetModulePath(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(tempDir string) // Test setup
		expected      string
		expectFailure bool
	}{
		{
			name: "Valid go.mod in current directory",
			setup: func(tempDir string) {
				writeGoMod(tempDir, "module github.com/test/module")
			},
			expected:      "github.com/test/module",
			expectFailure: false,
		},
		{
			name: "Valid go.mod in parent directory",
			setup: func(tempDir string) {
				parentDir := filepath.Join(tempDir, "parent")
				subDir := filepath.Join(parentDir, "subdir")

				_ = os.MkdirAll(subDir, 0755)
				writeGoMod(parentDir, "module github.com/test/parentmodule")

				_ = os.Chdir(subDir)
			},
			expected:      "github.com/test/parentmodule",
			expectFailure: false,
		},
		{
			name: "No go.mod, fallback to go list -m",
			setup: func(tempDir string) {
				// mocked `go list -m`
				core.GetModulePathFromGoList = func() (string, error) {
					return "github.com/fallback/module", nil
				}
			},
			expected:      "github.com/fallback/module",
			expectFailure: false,
		},
		{
			name: "Neither go.mod nor go list -m works",
			setup: func(tempDir string) {
				// fail `go list -m`
				core.GetModulePathFromGoList = func() (string, error) {
					return "", os.ErrNotExist
				}
			},
			expected:      "",
			expectFailure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "testmodule")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			originalWD, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			defer os.Chdir(originalWD)

			tt.setup(tempDir)

			result, err := core.GetModulePath()

			if (err != nil) != tt.expectFailure {
				t.Errorf("Expected failure: %v, but got error: %v", tt.expectFailure, err)
			}
			if result != tt.expected {
				t.Errorf("Expected: %q, but got: %q", tt.expected, result)
			}
		})
	}
}

// writeGoMod writes a fake go.mod file to the specified directory
func writeGoMod(dir, content string) {
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644)
}

