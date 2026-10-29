package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/imports"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/linesafe"
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

	formatted, err := formatSource(filename, src)
	if err != nil {
		return err
	}

	sorted, err := FormatImports(formatted, opt)
	if err != nil {
		return err
	}

	if opt.WriteFlag {
		if bytes.Equal(sorted, src) {
			// Already formatted: the file is left alone (not even its
			// modification time changes).
			return nil
		}
		return writeFileAtomic(filename, sorted)
	} else {
		_, err := os.Stdout.Write(sorted)
		return err
	}
}

// formatSource formats src as gofmt does and merges its import
// declarations into one (imports.Process with FormatOnly). A source that
// contains a line directive goes to linesafe, a copy of that code that
// ignores line directives where x/tools applied them: x/tools v0.52.0 loops
// forever, printing "panic avoided", or panics on some of those sources.
func formatSource(filename string, src []byte) ([]byte, error) {
	if linesafe.HasLineDirective(src) {
		return linesafe.Process(filename, src)
	}
	return imports.Process(filename, src, &imports.Options{
		FormatOnly: true,
		Comments:   true,
	})
}

// writeTempFile writes the new content into the temporary file; tests
// replace it to inject a failure in the middle of the write.
var writeTempFile = func(f *os.File, data []byte) error {
	_, err := f.Write(data)
	return err
}

// writeFileAtomic replaces the content of filename with data, or leaves the
// file as it was. The data goes to a temporary file in the directory of the
// file, which is synced and then renamed over the file: a failure at any
// point (disk full, file size limit) leaves the original file intact and
// removes the temporary file. A symbolic link stays a link: the file it
// resolves to is replaced. The file keeps its permission bits.
func writeFileAtomic(filename string, data []byte) (err error) {
	target, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return err
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: not a regular file", filename)
	}
	// The rename below needs only the directory to be writable. A file the
	// user may not write (a read-only file) is refused as when it was written
	// in place, before anything is created.
	f, err := os.OpenFile(target, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".goreg-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	closed := false
	defer func() {
		if err != nil {
			if !closed {
				_ = tmp.Close()
			}
			_ = os.Remove(tmpName)
		}
	}()

	if err = writeTempFile(tmp, data); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	closed = true
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmpName, target)
}
