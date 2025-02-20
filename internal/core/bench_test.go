package core_test

import (
	"testing"

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
	modulePath := "github.com/test/project"

	for i := 0; i < b.N; i++ {
		_, err := core.FormatImports(sampleGoCode, modulePath)
		if err != nil {
			b.Fatalf("FormatImports failed: %v", err)
		}
	}
}
