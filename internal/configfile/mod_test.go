package configfile

import (
	"os"
	"path/filepath"
	"testing"
)

// createTempTomlFile creates a temporary TOML file for testing.
func createTempTomlFile(t *testing.T, dir, content string) string {
	t.Helper()

	tomlPath := filepath.Join(dir, "goreg.toml")
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp TOML file: %v", err)
	}
	return tomlPath
}

// TestFindGoregToml checks if the function correctly finds the nearest goreg.toml.
func TestFindGoregToml(t *testing.T) {
	tempDir := t.TempDir() // Create a temporary directory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create goreg.toml in tempDir
	expectedTomlPath := createTempTomlFile(t, tempDir, "key = \"value\"")

	// Normalize expected path to resolve symlinks (fix for macOS /private issue)
	normalizedExpectedPath, err := filepath.EvalSymlinks(expectedTomlPath)
	if err != nil {
		t.Fatalf("failed to normalize expected path: %v", err)
	}

	// Change working directory to subDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(origDir) // Restore original directory after test

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test finding goreg.toml
	foundPath, err := findGoregToml()
	if err != nil {
		t.Fatalf("findGoregToml failed: %v", err)
	}

	// Normalize found path
	normalizedFoundPath, err := filepath.EvalSymlinks(foundPath)
	if err != nil {
		t.Fatalf("failed to normalize found path: %v", err)
	}

	// Compare normalized paths
	if normalizedFoundPath != normalizedExpectedPath {
		t.Errorf("expected %s, got %s", normalizedExpectedPath, normalizedFoundPath)
	}
}

// TestLoadToml verifies that the TOML file is correctly parsed into a map.
func TestLoadToml(t *testing.T) {
	tempDir := t.TempDir()
	tomlPath := createTempTomlFile(t, tempDir, `
		[key]
		subkey = "hello"
		number = 42
		boolValue = true
	`)

	config, err := loadToml(tomlPath)
	if err != nil {
		t.Fatalf("failed to load TOML: %v", err)
	}

	// Check parsed values
	key, ok := config["key"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected key to be a map, got %T", config["key"])
	}

	if key["subkey"] != "hello" {
		t.Errorf("expected key.subkey to be 'hello', got %v", key["subkey"])
	}

	if key["number"] != int64(42) { // go-toml/v2 decodes numbers as int64
		t.Errorf("expected key.number to be 42, got %v", key["number"])
	}

	if key["boolValue"] != true {
		t.Errorf("expected key.boolValue to be true, got %v", key["boolValue"])
	}
}
