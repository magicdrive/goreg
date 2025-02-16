package core

import (
	"os"

	"golang.org/x/tools/imports"
)

func Apply(filename string, _modulePath string, writeToFile bool) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	ModulePath = _modulePath

	formatted, err := imports.Process(filename, src, &imports.Options{
		FormatOnly: true,
		Comments:   true,
	})
	if err != nil {
		return err
	}

	sorted, err := FormatImports(formatted)
	if err != nil {
		return err
	}

	if writeToFile {
		return os.WriteFile(filename, sorted, 0644)
	} else {
		_, err := os.Stdout.Write(sorted)
		return err
	}
}
