package core

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Line directives (//line file:line, /*line file:line*/) change the line
// numbers go/token reports. golang.org/x/tools/imports v0.52.0 mixes them
// with the lines of the file: on the sources below it printed
// "panic avoided: ..." forever (enough to fill a disk when the output went
// to a log) or panicked. goreg formats such sources with internal/linesafe.

// logSink keeps the first bytes the standard logger writes (x/tools logs
// "panic avoided" through it) and drops the rest, so that a regression
// cannot flood the test output.
type logSink struct {
	mu  sync.Mutex
	buf []byte
}

func (s *logSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if room := 4096 - len(s.buf); room > 0 {
		s.buf = append(s.buf, p[:min(len(p), room)]...)
	}
	return len(p), nil
}

func (s *logSink) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return string(s.buf)
}

// applyWithin runs Apply with -w on a file holding src. It fails the test
// when Apply does not return within 10 seconds or logs anything. The
// logger stays redirected after a timeout: the goroutine may still run.
func applyWithin(t *testing.T, src string) (string, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src, 0o644)
	sink := &logSink{}
	log.SetOutput(sink)
	done := make(chan error, 1)
	go func() { done <- Apply(writeOpt(path)) }()
	select {
	case err := <-done:
		log.SetOutput(os.Stderr)
		if logged := sink.String(); logged != "" {
			t.Errorf("Apply logged:\n%s", logged)
		}
		return readFile(t, path), err
	case <-time.After(10 * time.Second):
		t.Fatalf("Apply did not return within 10s; it logged:\n%.500s", sink.String())
		return "", nil
	}
}

func TestApply_LineDirectives(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{
			"directive before the block (x/tools looped forever)",
			"package main\n\n//line foo.go:100\nimport (\n\t// doc for os\n\t\"os\"\n\t\"fmt\" // end\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
			"package main\n\n//line foo.go:100\nimport (\n\t// doc for os\n\t\"fmt\" // end\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
		},
		{
			"block directive (x/tools looped forever)",
			"package main\n\n/*line foo.go:100:1*/ import (\n\t\"os\"\n\t\"fmt\" // end\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
			// gofmt puts the declaration on its own line, as without goreg.
			"package main\n\n/*line foo.go:100:1*/\nimport (\n\t\"fmt\" // end\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
		},
		{
			"duplicate import after a directive (x/tools panicked)",
			"package main\n\n//line foo.go:100\nimport (\n\t\"os\"\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
			"package main\n\n//line foo.go:100\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
		},
		{
			// v1.3.12 dropped the directive and the "// end" comment, and
			// indented a directive it kept, which made it a plain comment.
			"directive inside the block stays a directive",
			"package main\n\nimport (\n\t\"os\"\n//line foo.go:1\n\t\"fmt\" // end\n\t\"bytes\"\n)\n\nvar _ = bytes.MinRead\n\nfunc main() { fmt.Println(os.Args) }\n",
			"package main\n\nimport (\n//line foo.go:1\n\t\"bytes\"\n\t\"fmt\" // end\n\t\"os\"\n)\n\nvar _ = bytes.MinRead\n\nfunc main() { fmt.Println(os.Args) }\n",
		},
		{
			"indented //line is a plain comment and stays indented",
			"package main\n\nimport (\n\t\"os\"\n\t//line foo.go:1\n\t\"fmt\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
			"package main\n\nimport (\n\t//line foo.go:1\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
		},
		{
			// The comment of "os" stays with "os"; with directive lines it
			// went to the line comment map entry of another import.
			"directive after a commented import",
			"package main\n\nimport (\n\t\"os\" // keep\n//line x.go:4\n\t\"fmt\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
			"package main\n\nimport (\n//line x.go:4\n\t\"fmt\"\n\t\"os\" // keep\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
		},
		{
			"comment before an import on its line, after a directive",
			"package main\n\n//line x.go:50\nimport (\n\t/* c */ \"fmt\"\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
			"package main\n\n//line x.go:50\nimport (\n\t/* c */\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n",
		},
		{
			"directive after the imports",
			"package main\n\nimport (\n\t\"os\"\n\t\"fmt\" // end\n)\n\n//line foo.go:100\nfunc main() { fmt.Println(os.Args) }\n",
			"package main\n\nimport (\n\t\"fmt\" // end\n\t\"os\"\n)\n\n//line foo.go:100\nfunc main() { fmt.Println(os.Args) }\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := applyWithin(t, tc.src)
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}
			if got != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.want)
			}
			// A formatted file is left as it is.
			again, err := applyWithin(t, got)
			if err != nil || again != got {
				t.Errorf("second run changed the file (err %v):\n%s", err, again)
			}
		})
	}
}

// A directive the parser rejects is an error, and -w leaves the file alone.
func TestApply_InvalidLineDirective(t *testing.T) {
	src := "package main\n\n//line foo.go:0\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n"
	got, err := applyWithin(t, src)
	if err == nil || !strings.Contains(err.Error(), "invalid line number") {
		t.Errorf("err = %v, want an invalid line number error", err)
	}
	if got != src {
		t.Errorf("the file changed:\n%s", got)
	}
}
