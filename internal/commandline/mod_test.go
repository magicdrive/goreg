package commandline_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/model"
)

func TestOptParse(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *commandline.Option
		wantErr  bool
	}{
		{
			name: "No options (default values)",
			args: []string{},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Single file argument",
			args: []string{"main.go"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "main.go",
			},
			wantErr: false,
		},
		{
			name: "Specify organization name",
			args: []string{"--organization", "github.com/myorg"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "github.com/myorg",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Short flag for organization name",
			args: []string{"-n", "github.com/myorg"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "github.com/myorg",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Specify order",
			args: []string{"--order", "std,local,thirdparty,organization"},
			expected: &commandline.Option{
				ImportOrder:          []model.ImportGroup{model.StdLib, model.Local, model.ThirdParty, model.Organization},
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Short flag for order",
			args: []string{"-o", "std,local,thirdparty,organization"},
			expected: &commandline.Option{
				ImportOrder:          []model.ImportGroup{model.StdLib, model.Local, model.ThirdParty, model.Organization},
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name:     "Error order",
			args:     []string{"--order", "std,local,organization"},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Specify order other name 1.",
			args: []string{"-o", "s,t,o,l"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Specify order other name 2.",
			args: []string{"-o", "stdlib,3rd,org,local"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Specify order other name 3.",
			args: []string{"-o", "s,3,org,local"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Specify order other name 4.",
			args: []string{"-o", "s,3rd_party,org,local"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Specify order other name 5.",
			args: []string{"-o", "s,third_party,org,local"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Enable minimize group flag",
			args: []string{"--minimize-group"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    true,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Enable sort include alias flag",
			args: []string{"--sort-include-alias"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: true,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Set module path",
			args: []string{"--local", "myproject/module"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "myproject/module",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Enable write flag",
			args: []string{"--write"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            true,
				HelpFlag:             false,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Enable help flag",
			args: []string{"--help"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             true,
				VersionFlag:          false,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
		{
			name: "Enable version flag",
			args: []string{"--version"},
			expected: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				OrganizationName:     "",
				MinimizeGroupFlag:    false,
				SortIncludeAliasFlag: false,
				WriteFlag:            false,
				HelpFlag:             false,
				VersionFlag:          true,
				ModulePath:           "",
				FileName:             "",
			},
			wantErr: false,
		},
	}

	os.Setenv("GOREG_NOT_USE_CONFIGFILE", "1")
	defer os.Setenv("GOREG_NOT_USE_CONFIGFILE", "")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := commandline.OptParse(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error status: got %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if got.ImportOrder != nil && len(got.ImportOrder) != len(tt.expected.ImportOrder) {
					t.Errorf("ImportOrder mismatch: got %v, want %v", got.ImportOrder, tt.expected.ImportOrder)
				}

				got.FlagSet = nil
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("expected %+v, got %+v", tt.expected, got)
				}

			}
		})
	}
}

// inConfigDir runs the test in a new directory holding a goreg.toml with
// the given content, with no configuration file found in the home directory.
func inConfigDir(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "goreg.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("GOREG_NOT_USE_CONFIGFILE", "")
	t.Chdir(dir)
}

func TestOptParse_RemoveImportComment(t *testing.T) {
	t.Setenv("GOREG_NOT_USE_CONFIGFILE", "1")
	for _, args := range [][]string{{"-r"}, {"--remove-import-comment"}} {
		_, got, err := commandline.OptParse(args)
		if err != nil {
			t.Fatal(err)
		}
		if !got.RemoveImportCommentFlag {
			t.Errorf("%v: RemoveImportCommentFlag is false", args)
		}
	}
	_, got, err := commandline.OptParse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.RemoveImportCommentFlag {
		t.Error("RemoveImportCommentFlag is true by default")
	}
}

func TestOptParse_RemoveImportCommentFromConfig(t *testing.T) {
	inConfigDir(t, "[format]\nremove_import_comment = true\n")
	_, got, err := commandline.OptParse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !got.RemoveImportCommentFlag {
		t.Error("remove_import_comment = true in goreg.toml did not set RemoveImportCommentFlag")
	}
}

// An empty goreg.toml is the default configuration.
func TestOptParse_EmptyConfig(t *testing.T) {
	inConfigDir(t, "")
	_, got, err := commandline.OptParse([]string{"main.go"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.ImportOrder, model.DefaultOrder) {
		t.Errorf("ImportOrder %v, want %v", got.ImportOrder, model.DefaultOrder)
	}
	if got.FileName != "main.go" {
		t.Errorf("FileName %q", got.FileName)
	}
}

// A goreg.toml that is not valid TOML, or cannot be read, is an error
// (never a panic, never the defaults).
func TestOptParse_BadConfig(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		inConfigDir(t, "[import\n")
		_, got, err := commandline.OptParse([]string{"main.go"})
		if err == nil {
			t.Fatalf("no error, got %+v", got)
		}
		if !strings.Contains(err.Error(), "goreg.toml") {
			t.Errorf("the error does not name the file: %v", err)
		}
	})
	t.Run("unreadable", func(t *testing.T) {
		dir := t.TempDir()
		// A directory named goreg.toml is found but cannot be read.
		if err := os.Mkdir(filepath.Join(dir, "goreg.toml"), 0o755); err != nil {
			t.Fatal(err)
		}
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		t.Setenv("GOREG_NOT_USE_CONFIGFILE", "")
		t.Chdir(dir)
		if _, got, err := commandline.OptParse([]string{"main.go"}); err == nil {
			t.Fatalf("no error, got %+v", got)
		}
	})
}
