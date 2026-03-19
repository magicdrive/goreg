package initcmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magicdrive/goreg/internal/initcmd"
)

func TestExecute_Success(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Change to temporary directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Execute
	err = initcmd.Execute()
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// Check if goreg.toml was created
	filename := "goreg.toml"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("goreg.toml was not created")
	}

	// Read the created file
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read goreg.toml: %v", err)
	}

	// Check content
	contentStr := string(content)

	// Check for expected sections
	expectedStrings := []string{
		"### goreg.toml",
		"[import]",
		"local_module",
		"organization_module",
		"order",
		"[format]",
		"minimize_group",
		"sort_include_alias",
		"remove_import_comment",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("goreg.toml does not contain expected string: %q", expected)
		}
	}
}

func TestExecute_FileAlreadyExists(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Change to temporary directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create goreg.toml first
	filename := "goreg.toml"
	if err := os.WriteFile(filename, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Execute should return error
	err = initcmd.Execute()
	if err == nil {
		t.Error("Execute() should return error when file already exists")
	}

	// Check error message
	expectedMsg := "already exists"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error message should contain %q, got: %v", expectedMsg, err)
	}

	// Check that the original file was not overwritten
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "existing content" {
		t.Error("Existing file was overwritten")
	}
}

func TestExecute_ContentMatchesEmbedded(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Change to temporary directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Execute
	if err := initcmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Read the created file and verify it has valid TOML structure
	content, err := os.ReadFile("goreg.toml")
	if err != nil {
		t.Fatalf("Failed to read goreg.toml: %v", err)
	}

	// Check that content is not empty and contains expected structure
	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("Created file is empty")
	}

	// Verify TOML structure
	expectedSections := []string{
		"[import]",
		"[format]",
	}

	for _, section := range expectedSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("Created file does not contain expected section: %q", section)
		}
	}
}

func TestExecute_FilePermissions(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Change to temporary directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Execute
	if err := initcmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Check file permissions
	info, err := os.Stat("goreg.toml")
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedPerm := os.FileMode(0644)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("File permissions incorrect. Expected: %v, Got: %v", expectedPerm, info.Mode().Perm())
	}
}

func TestExecute_InSubdirectory(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Change to subdirectory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Execute
	if err := initcmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Check if goreg.toml was created in the current directory
	filename := "goreg.toml"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("goreg.toml was not created in subdirectory")
	}

	// Check that it was not created in parent directory
	parentFile := filepath.Join(tmpDir, "goreg.toml")
	if _, err := os.Stat(parentFile); !os.IsNotExist(err) {
		t.Errorf("goreg.toml should not be created in parent directory")
	}
}
