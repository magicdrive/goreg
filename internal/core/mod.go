package core

import (
	"os"
	"path/filepath"

	"golang.org/x/tools/imports"

	"github.com/magicdrive/goreg/internal/commandline"
)

func Apply(opt *commandline.Option) error {
	filename := opt.FileName
	basename := filepath.Base(filename)
	if basename == "go.mod" || basename == "go.sum" {
		return nil
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	formatted, err := imports.Process(filename, src, &imports.Options{
		FormatOnly: true,
		Comments:   true,
	})
	if err != nil {
		return err
	}

	sorted, err := FormatImports(formatted, opt)
	if err != nil {
		return err
	}

	if opt.WriteFlag {
		return os.WriteFile(filename, sorted, 0644)
	} else {
		_, err := os.Stdout.Write(sorted)
		return err
	}
}
