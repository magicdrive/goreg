package core

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/model"
)

const unsortedSrc = `package main

import (
	"os"
	"fmt"
)

func main() { fmt.Println(os.Args) }
`

const sortedSrc = `package main

import (
	"fmt"
	"os"
)

func main() { fmt.Println(os.Args) }
`

func writeOpt(filename string) *commandline.Option {
	return &commandline.Option{
		ImportOrder: model.DefaultOrder,
		ModulePath:  "example.com/m",
		WriteFlag:   true,
		FileName:    filename,
	}
}

func writeFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
	// The umask does not apply to Chmod.
	if err := os.Chmod(path, perm); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// dirEntries lists the names in dir, to find leftover temporary files.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestApplyWrite_ReplacesContentAndKeepsMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, path, unsortedSrc, 0o600)

	if err := Apply(writeOpt(path)); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != sortedSrc {
		t.Errorf("content:\n%s", got)
	}
	if names := dirEntries(t, dir); len(names) != 1 {
		t.Errorf("files left in the directory: %v", names)
	}
	if runtime.GOOS == "windows" {
		// Windows has no Unix permission bits (only read-only).
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode %v, want 0600", info.Mode().Perm())
	}
}

// A failure in the middle of the write leaves the original file intact and
// no temporary file behind.
func TestApplyWrite_FailureLeavesFileIntact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, path, unsortedSrc, 0o644)

	injected := errors.New("injected: file too large")
	orig := writeTempFile
	writeTempFile = func(f *os.File, data []byte) error {
		if _, err := f.Write(data[:len(data)/2]); err != nil {
			return err
		}
		return injected
	}
	t.Cleanup(func() { writeTempFile = orig })

	if err := Apply(writeOpt(path)); !errors.Is(err, injected) {
		t.Fatalf("Apply: %v, want the injected error", err)
	}
	if got := readFile(t, path); got != unsortedSrc {
		t.Errorf("the original file changed:\n%s", got)
	}
	if names := dirEntries(t, dir); len(names) != 1 {
		t.Errorf("files left in the directory: %v", names)
	}
}

// When the temporary file cannot be created (a read-only directory), the
// file is not written and stays as it was.
func TestApplyWrite_ReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the read-only attribute of a Windows directory does not prevent creating files in it")
	}
	if os.Geteuid() == 0 {
		t.Skip("root writes into read-only directories")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, path, unsortedSrc, 0o644)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if err := Apply(writeOpt(path)); err == nil {
		t.Fatal("Apply succeeded in a read-only directory")
	}
	if got := readFile(t, path); got != unsortedSrc {
		t.Errorf("the original file changed:\n%s", got)
	}
}

// A symbolic link stays a link; the file it points to gets the new content.
func TestApplyWrite_SymlinkKeepsLink(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(realDir, "main.go")
	writeFile(t, target, unsortedSrc, 0o640)
	link := filepath.Join(dir, "link.go")
	if err := os.Symlink(filepath.Join("real", "main.go"), link); err != nil {
		t.Skipf("cannot create a symbolic link here (Windows without the privilege): %v", err)
	}

	if err := Apply(writeOpt(link)); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is no longer a symbolic link (mode %v)", link, fi.Mode())
	}
	if got := readFile(t, target); got != sortedSrc {
		t.Errorf("target content:\n%s", got)
	}
	if names := dirEntries(t, realDir); len(names) != 1 {
		t.Errorf("files left next to the target: %v", names)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o640 {
			t.Errorf("target mode %v, want 0640", info.Mode().Perm())
		}
	}
}

// An already formatted file is not written at all (its modification time
// does not change): the write function is never called.
func TestApplyWrite_UnchangedIsNotWritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, path, sortedSrc, 0o644)

	orig := writeTempFile
	writeTempFile = func(*os.File, []byte) error {
		t.Error("an unchanged file was written")
		return nil
	}
	t.Cleanup(func() { writeTempFile = orig })

	if err := Apply(writeOpt(path)); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != sortedSrc {
		t.Errorf("content changed:\n%s", got)
	}
}
