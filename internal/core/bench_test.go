package core_test

import (
	"testing"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/core"
)

var sampleGoCode = []byte(`package main

import (
	/* before fmt */ "fmt" /* after fmt */
	"github.com/pkg/errors" // after errors
	/* before log */ "log"
)

// main function
func main() {
	fmt.Println("Hello, world!")
}
`)

func BenchmarkFormatImports(b *testing.B) {
	opt := &commandline.Option{
		ImportOrder:          nil,
		OrganizationName:     "",
		MinimizeGroupFlag:    false,
		SortIncludeAliasFlag: false,
		WriteFlag:            false,
		HelpFlag:             false,
		VersionFlag:          false,
		ModulePath:           "github.com/test/project",
		FileName:             "",
		FlagSet:              nil,
	}

	for i := 0; i < b.N; i++ {
		_, err := core.FormatImports(sampleGoCode, opt)
		if err != nil {
			b.Fatalf("FormatImports failed: %v", err)
		}
	}
}
