package linesafe

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/tools/imports"
)

// Never pass a source with a line directive to imports.Process in these
// tests: x/tools v0.52.0 may loop forever on it, logging without bound.

// upstream is what goreg ran before: imports.Process with goreg's options.
func upstream(t *testing.T, src []byte) ([]byte, error) {
	t.Helper()
	if HasLineDirective(src) {
		t.Fatal("test bug: a source with a line directive must not reach imports.Process")
	}
	return imports.Process("x.go", src, &imports.Options{FormatOnly: true, Comments: true})
}

// process runs Process and fails the test when it does not return within
// 10 seconds (the goroutine is left behind; the test binary ends anyway).
func process(t *testing.T, src []byte) ([]byte, error) {
	t.Helper()
	type result struct {
		out []byte
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := Process("x.go", src)
		done <- result{out, err}
	}()
	select {
	case r := <-done:
		return r.out, r.err
	case <-time.After(10 * time.Second):
		t.Fatal("Process did not return within 10s")
		return nil, nil
	}
}

var directiveRe = regexp.MustCompile(`//line |/\*line `)

// neutral turns every line directive of src into a comment of the same
// length that is not a line directive ("//line " becomes "//li:e "), so that
// offsets and layout do not change. "//li:e" has the form of a directive
// (//word:word), which gofmt leaves as it is, as it leaves //line.
func neutral(t *testing.T, src string) []byte {
	t.Helper()
	if strings.Contains(src, "li:e ") {
		t.Fatal("test bug: the source already contains li:e")
	}
	return []byte(directiveRe.ReplaceAllStringFunc(src, func(m string) string {
		return strings.Replace(m, "line", "li:e", 1)
	}))
}

// oracle is the expected output for src: upstream's output for the neutral
// source, with the directives restored. The printer keeps a //line
// directive that starts a line at the start of the line (an ordinary
// comment in an import block is indented), so a directive in column 1 of
// src is unindented again.
func oracle(t *testing.T, src string) string {
	t.Helper()
	out, err := upstream(t, neutral(t, src))
	if err != nil {
		t.Fatalf("upstream on the neutral source: %v", err)
	}
	want := strings.ReplaceAll(string(out), "/*li:e ", "/*line ")
	col1 := map[string]bool{}
	for _, l := range strings.Split(src, "\n") {
		if strings.HasPrefix(l, "//line ") {
			col1[l] = true
		}
	}
	lines := strings.Split(want, "\n")
	for i, l := range lines {
		if text := strings.TrimLeft(l, " \t"); strings.HasPrefix(text, "//li:e ") {
			restored := "//line " + strings.TrimPrefix(text, "//li:e ")
			if col1[restored] {
				lines[i] = restored
			} else {
				lines[i] = strings.Replace(l, "//li:e ", "//line ", 1)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func TestProcess_LineDirectives(t *testing.T) {
	// A directive line below the line count of the file: x/tools then merged
	// the wrong lines instead of reaching its "panic avoided" branch.
	inRange := "package main\n//line x.go:20\nimport (\n\t\"os\" // o\n\t\"fmt\"\n)\n" +
		strings.Repeat("// a\nvar a = 1\n\n", 9) + "func main() { fmt.Println(os.Args) }\n"
	for _, tc := range []struct{ name, src string }{
		{"directive line within the line count", inRange},
		{"directive between groups", `package main

import (
	"os"
//line x.go:5
	"example.com/x"
)

var _ = x.X

func main() { fmt.Println(os.Args) }
`},
		{"directive before the block, comments in the block (looped forever)", `package main

//line foo.go:100
import (
	// doc for os
	"os"
	"fmt" // end
)

func main() { fmt.Println(os.Args) }
`},
		{"duplicate import after a directive (panicked)", `package main

//line foo.go:100
import (
	"os"
	"fmt"
	"os"
)

func main() { fmt.Println(os.Args) }
`},
		{"block directive before the declaration (looped forever)", `package main

/*line foo.go:100:1*/ import (
	"os"
	"fmt" // end
)

func main() { fmt.Println(os.Args) }
`},
		{"directive lowering the line inside the block", `package main

import (
	"os"
//line foo.go:1
	"fmt" // end
	"bytes"
)

var _ = bytes.MinRead

func main() { fmt.Println(os.Args) }
`},
		{"directive raising the line inside the block", `package main

import (
	"os"
//line gen.go:50
	"fmt"
)

func main() { fmt.Println(os.Args) }
`},
		{"several directives", `package main

//line a.go:500
import (
	"strings" // s
	"os"
//line b.go:3
	"fmt" // f
	"bytes"
)

//line c.go:9000
var _, _ = bytes.MinRead, strings.Clone

func main() { fmt.Println(os.Args) }
`},
		{"directive after the imports", `package main

import (
	"os"
	"fmt" // end
)

//line foo.go:100
func main() { fmt.Println(os.Args) }
`},
		{"golang/go#51916 input", `package main
//line mah.go:600

import (
"foo.com/a.thing"
"foo.com/surprise"
"foo.com/v1"
"foo.com/other/v2"
"foo.com/other/v3"
"foo.com/go-thing"
"foo.com/go-wrong"
)

var _ = []interface{}{bar.X, v1.Y, a.A, v2.V2, other.V3, thing.Thing, gow.Wrong}
`},
		{"filename with a colon, spaces and non-ASCII", `package main

//line C:\work dir\gen 生成.go:1073741824
import (
	"os"
	"fmt" // end
)

func main() { fmt.Println(os.Args) }
`},
		{"line and column", `package main

//line foo.go:100:7
import (
	"os"
	"fmt" // end
)

func main() { fmt.Println(os.Args) }
`},
		{"no filename", `package main

//line :100
import (
	"os"
	"fmt" // end
)

func main() { fmt.Println(os.Args) }
`},
		{"two import declarations", `package main

//line foo.go:100
import "os"

import (
	"fmt" // end
	"bytes"
)

var _ = bytes.MinRead

func main() { fmt.Println(os.Args) }
`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !HasLineDirective([]byte(tc.src)) {
				t.Fatal("test bug: no directive")
			}
			got, err := process(t, []byte(tc.src))
			if err != nil {
				t.Fatal(err)
			}
			if want := oracle(t, tc.src); string(got) != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

// A directive the parser rejects is an error, as with imports.Process (whose
// parser rejects it before any import is sorted).
func TestProcess_InvalidDirective(t *testing.T) {
	for _, line := range []string{"0", "1073741825", "-1"} {
		src := "package main\n\n//line foo.go:" + line + "\nimport (\n\t\"os\"\n\t\"fmt\" // end\n)\n"
		if _, err := process(t, []byte(src)); err == nil || !strings.Contains(err.Error(), "invalid line number") {
			t.Errorf("line %s: err = %v, want an invalid line number error", line, err)
		}
	}
}

// An indented "//line" is an ordinary comment: it is not moved to column 1.
func TestProcess_IndentedLineCommentStaysIndented(t *testing.T) {
	src := "package main\n\nimport (\n\t\"os\"\n\t//line foo.go:1\n\t\"fmt\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n"
	got, err := process(t, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "\n\t//line foo.go:1\n") {
		t.Errorf("the comment lost its indentation:\n%s", got)
	}
}

// Without a line directive, Process is imports.Process, byte for byte.
func TestProcess_SameAsUpstreamWithoutDirectives(t *testing.T) {
	cases := map[string]string{
		"no imports":   "package p\n",
		"empty block":  "package p\n\nimport ()\n",
		"single":       "package p\n\nimport \"os\"\n\nvar _ = os.Args\n",
		"unsorted":     "package p\n\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprint(os.Args)\n",
		"groups":       "package p\n\nimport (\n\t\"os\"\n\t\"example.com/x\"\n\t\"fmt\"\n\n\t\"golang.org/x/tools\"\n\t\"bytes\"\n)\n",
		"comments":     "package p\n\nimport (\n\t// doc\n\t\"os\" // os\n\t/* a */ \"fmt\" /* b */\n\t\"bytes\" // b\n)\n",
		"duplicates":   "package p\n\nimport (\n\t\"os\"\n\t\"fmt\"\n\t\"os\"\n\tf \"fmt\"\n\t_ \"fmt\"\n\t. \"fmt\"\n)\n",
		"two decls":    "package p\n\nimport \"os\"\n\nimport (\n\t\"fmt\"\n\t\"bytes\"\n)\n",
		"cgo":          "package p\n\n// #include <stdio.h>\nimport \"C\"\n\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n",
		"appengine":    "package p\n\nimport (\n\t\"appengine/x\"\n\t\"os\"\n\t\"a.b/c\"\n)\n",
		"crlf":         "package p\r\n\r\nimport (\r\n\t\"os\"\r\n\t\"fmt\"\r\n)\r\n",
		"bom":          "\xef\xbb\xbfpackage p\n\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n",
		"blank lines":  "package p\n\nimport (\n\n\t\"os\"\n\n\n\t\"fmt\"\n\n)\n",
		"not go":       "this is not Go\n",
		"line comment": "package p\n\n// line foo.go:100 is not a directive\nimport (\n\t\"os\"\n\t\"fmt\" // x\n)\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) { compareWithUpstream(t, []byte(src)) })
	}
}

func compareWithUpstream(t *testing.T, src []byte) {
	t.Helper()
	want, werr := upstream(t, src)
	got, gerr := process(t, src)
	if (werr == nil) != (gerr == nil) || (werr != nil && werr.Error() != gerr.Error()) {
		t.Fatalf("errors differ: got %v, imports.Process %v", gerr, werr)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output differs from imports.Process:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// The same comparison over real code: the Go files of some standard library
// packages and of goreg. Files with a line directive (or the text of one)
// are skipped, as they must never reach imports.Process.
func TestProcess_SameAsUpstreamOnRealCode(t *testing.T) {
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Fatalf("go env GOROOT: %v", err)
	}
	goroot := strings.TrimSpace(string(out))
	roots := []string{"../.."}
	for _, p := range []string{"go", "fmt", "net/http", "strings", "os", "cmd/gofmt", "encoding/json"} {
		roots = append(roots, filepath.Join(goroot, "src", filepath.FromSlash(p)))
	}
	n := 0
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			t.Logf("skipping %s: %v", root, err)
			continue
		}
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".")) && path != root {
				return filepath.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if HasLineDirective(src) {
				return nil
			}
			n++
			t.Run(path, func(t *testing.T) { compareWithUpstream(t, src) })
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if n < 500 {
		t.Fatalf("compared only %d files", n)
	}
	t.Logf("compared %d files", n)
}

func TestHasLineDirective(t *testing.T) {
	for src, want := range map[string]bool{
		"package p\n":                      false,
		"package p\n//line a.go:1\n":       true,
		"package p\n/*line a.go:1:1*/\n":   true,
		"package p\n// line a.go:1\n":      false,
		"package p\nvar s = \"//line \"\n": true, // over-approximation, harmless
	} {
		if got := HasLineDirective([]byte(src)); got != want {
			t.Errorf("HasLineDirective(%q) = %v, want %v", src, got, want)
		}
	}
}

// copiedFrom is the x/tools version this package was copied from. When
// go.mod moves to another version, check whether x/tools now handles line
// directives (internal/imports/sortimports.go): if it does, remove this
// package and call imports.Process for every source; if not, compare the
// copy with the new version and update copiedFrom.
const copiedFrom = "v0.52.0"

func TestCopiedFromTheRequiredVersion(t *testing.T) {
	gomod, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^\s*golang\.org/x/tools (\S+)`).FindSubmatch(gomod)
	if m == nil {
		t.Fatal("go.mod does not require golang.org/x/tools")
	}
	if got := string(m[1]); got != copiedFrom {
		t.Errorf("go.mod requires golang.org/x/tools %s, linesafe is a copy of %s: see copiedFrom", got, copiedFrom)
	}
}
