package commandline_test

import (
	"os"
	"reflect"
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
