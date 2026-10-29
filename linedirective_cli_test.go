package main

// The goreg binary on a source with a line directive before its imports:
// up to v1.3.12 it printed "panic avoided: ..." to stderr until it was
// killed. The run is limited in time and output, so a regression
// fails the test instead of filling the disk.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// cappedBuffer keeps at most max bytes; a write beyond them fails, which
// makes os/exec stop copying and close the pipe.
type cappedBuffer struct {
	mu  sync.Mutex
	buf []byte
	max int
}

var errOutputLimit = errors.New("output limit reached")

func (b *cappedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.buf)+len(p) > b.max {
		b.buf = append(b.buf, p[:b.max-len(b.buf)]...)
		return 0, errOutputLimit
	}
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *cappedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}

func TestCLI_LineDirectiveTerminates(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "goreg")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	src := "package main\n\n//line foo.go:100\nimport (\n\t// doc for os\n\t\"os\"\n\t\"fmt\" // end\n)\n\nfunc main() { fmt.Println(os.Args) }\n"
	file := filepath.Join(dir, "line.go")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"--local", "example.com/m", file},
		{"--local", "example.com/m", "-w", file},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		cmd := exec.CommandContext(ctx, bin, args...)
		cmd.Env = append(os.Environ(), "GOREG_NOT_USE_CONFIGFILE=1")
		cmd.WaitDelay = 5 * time.Second
		out := &cappedBuffer{max: 64 << 10}
		cmd.Stdout, cmd.Stderr = out, out
		err := cmd.Run()
		timedOut := ctx.Err() != nil
		cancel()
		if timedOut {
			t.Fatalf("goreg %s did not finish within 20s; output starts:\n%.500s", strings.Join(args, " "), out)
		}
		if err != nil {
			t.Fatalf("goreg %s: %v\n%.2000s", strings.Join(args, " "), err, out)
		}
		if strings.Contains(out.String(), "panic avoided") {
			t.Fatalf("goreg %s printed the x/tools diagnostic:\n%.500s", strings.Join(args, " "), out)
		}
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	want := "package main\n\n//line foo.go:100\nimport (\n\t// doc for os\n\t\"fmt\" // end\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Args) }\n"
	if string(got) != want {
		t.Errorf("-w wrote:\n%s\nwant:\n%s", got, want)
	}
}
