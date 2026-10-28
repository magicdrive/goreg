package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGoregToml(t *testing.T) {
	tempDir := t.TempDir()

	tempFile := filepath.Join(tempDir, "goreg.toml")
	if err := os.WriteFile(tempFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp goreg.toml: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	foundPath, err := FindGoregToml()
	if err != nil {
		t.Fatalf("expected to find goreg.toml, but got error: %v", err)
	}

	expectedPath, _ := filepath.EvalSymlinks(tempFile)
	actualPath, _ := filepath.EvalSymlinks(foundPath)

	if actualPath != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, actualPath)
	}
}

func TestFindGoregToml_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	_, err = FindGoregToml()
	if err == nil {
		t.Fatal("expected error when goreg.toml is not found, but got nil")
	}
}

func TestLoadToml(t *testing.T) {
	tomlContent := `
[import]
local_module = "example_project"
organization_module = "github.com/example_org"
order = "std,thirdparty,organization,local"

[format]
minimize_group = true
sort_include_alias = false
remove_import_comment = true
`
	tempFile := filepath.Join(t.TempDir(), "goreg.toml")

	if err := os.WriteFile(tempFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to create temp goreg.toml: %v", err)
	}

	cfg, err := LoadToml(tempFile)
	if err != nil {
		t.Fatalf("failed to load toml: %v", err)
	}

	if cfg.Import.LocalModule != "example_project" {
		t.Errorf("expected local_module to be 'example_project', got %s", cfg.Import.LocalModule)
	}
	if cfg.Import.OrganizationModule != "github.com/example_org" {
		t.Errorf("expected organization_module to be 'github.com/example_org', got %s", cfg.Import.OrganizationModule)
	}
	if cfg.Import.Order != "std,thirdparty,organization,local" {
		t.Errorf("expected order to be 'std,thirdparty,organization,local', got %s", cfg.Import.Order)
	}
	if !cfg.Format.MinimizeGroup {
		t.Errorf("expected minimize_group to be true, got false")
	}
	if cfg.Format.SortIncludeAlias {
		t.Errorf("expected sort_include_alias to be false, got true")
	}
	if !cfg.Format.RemoveImportComment {
		t.Errorf("expected remove_import_comment to be true, got false")
	}
}

func TestLoadToml_Invalid(t *testing.T) {
	invalidToml := `
[import]
local_module = "example_project"
organization_module = "github.com/example_org"
order = "std,thirdparty,organization,local"

[format]
minimize_group = true
sort_include_alias = false
remove_import_comment = true
invalid_field == "error"  # ← TOML syntax error
`
	tempFile := filepath.Join(t.TempDir(), "goreg.toml")

	if err := os.WriteFile(tempFile, []byte(invalidToml), 0644); err != nil {
		t.Fatalf("failed to create temp invalid goreg.toml: %v", err)
	}

	_, err := LoadToml(tempFile)
	if err == nil {
		t.Fatal("expected error due to invalid TOML, but got nil")
	} else {
		t.Logf("expected error: %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Import.Order != "std,thirdparty,organization,local" {
		t.Errorf("expected default import order to be 'std,thirdparty,organization,local', got %s", cfg.Import.Order)
	}
}

// An empty goreg.toml, or one that sets only some keys, gets the defaults
// for the others.
func TestLoadToml_EmptyAndPartial(t *testing.T) {
	for _, content := range []string{"", "[format]\nminimize_group = true\n"} {
		tempFile := filepath.Join(t.TempDir(), "goreg.toml")
		if err := os.WriteFile(tempFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadToml(tempFile)
		if err != nil {
			t.Fatalf("%q: %v", content, err)
		}
		if cfg == nil {
			t.Fatalf("%q: nil configuration", content)
		}
		if cfg.Import.Order != "std,thirdparty,organization,local" {
			t.Errorf("%q: order %q, want the default", content, cfg.Import.Order)
		}
	}
}

// LoadConfig reports a goreg.toml it cannot load instead of dropping it.
func TestLoadConfig_Invalid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "goreg.toml"), []byte("[import\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOREG_NOT_USE_CONFIGFILE", "")
	t.Chdir(dir)
	cfg, err := LoadConfig()
	if err == nil {
		t.Fatalf("no error, got %+v", cfg)
	}
}
