package commandline_test

import (
	"reflect"
	"testing"

	"github.com/magicdrive/goreg/internal/commandline"
)

func TestOptParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		expect  commandline.Option
		wantErr bool
	}{
		{
			name:   "No arguments",
			args:   []string{},
			expect: commandline.Option{},
		},
		{
			name:   "Help flag",
			args:   []string{"--help"},
			expect: commandline.Option{HelpFlag: true},
		},
		{
			name:   "Version flag",
			args:   []string{"--version"},
			expect: commandline.Option{VersionFlag: true},
		},
		{
			name:   "Write flag",
			args:   []string{"-w"},
			expect: commandline.Option{WriteFlag: true},
		},
		{
			name:   "File name provided",
			args:   []string{"file.go"},
			expect: commandline.Option{FileName: "file.go"},
		},
		{
			name:   "Write flag and file name",
			args:   []string{"-w", "file.go"},
			expect: commandline.Option{WriteFlag: true, FileName: "file.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := commandline.OptParse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("OptParse() error = %v, wantErr %v", err, tt.wantErr)
			}
			got.FlagSet = nil // Ignore FlagSet pointer for comparison
			if !reflect.DeepEqual(*got, tt.expect) {
				t.Errorf("OptParse() got = %+v, want %+v", *got, tt.expect)
			}
		})
	}
}
