package core_test

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/core"
	"github.com/magicdrive/goreg/internal/model"
)

func TestFormatImports(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
		opt      *commandline.Option
	}{
		{
			name: "Standard case",
			input: `package main

import (
	"fmt"
	"github.com/pkg/errors"
	"myproject/module"
)
`,
			expected: `package main

import (
	"fmt"

	"github.com/pkg/errors"

	"myproject/module"
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder: model.DefaultOrder,
				ModulePath:  "myproject/module",
			},
		},
		{
			name: "Unsorted imports",
			input: `package main

import (
	"myproject/module"
	"fmt"
	"github.com/pkg/errors"
)
`,
			expected: `package main

import (
	"fmt"

	"github.com/pkg/errors"

	"myproject/module"
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder: model.DefaultOrder,
				ModulePath:  "myproject/module",
			},
		},

		{
			name: "MinimizeGroupFlag enabled",
			input: `package main

import (
	"fmt"

	mylog "log"

	"github.com/pkg/errors"

	"myproject/module"
)
`,
			expected: `package main

import (
	"fmt"
	mylog "log"

	"github.com/pkg/errors"

	"myproject/module"
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder:       model.DefaultOrder,
				MinimizeGroupFlag: true,
				ModulePath:        "myproject/module",
			},
		},
		{
			name: "SortIncludeAliasFlag enabled",
			input: `package main

import (
	cmd "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"fmt"
)
`,
			expected: `package main

import (
	"fmt"

	cmd "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder:          model.DefaultOrder,
				SortIncludeAliasFlag: true,
				ModulePath:           "myproject/module",
			},
		},
		{
			name: "OrganizationName specified",
			input: `package main

import (
	"fmt"
	"github.com/pkg/errors"
	"orgname/project/module"
)
`,
			expected: `package main

import (
	"fmt"

	"github.com/pkg/errors"

	"orgname/project/module"
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder:      model.DefaultOrder,
				OrganizationName: "orgname",
				ModulePath:       "myproject/module",
			},
		},
		{
			name: "Local module mixed with third-party",
			input: `package main

import (
	"myproject/module"
	"fmt"
	"github.com/pkg/errors"
	"myproject/module/utils"
)
`,
			expected: `package main

import (
	"fmt"

	"github.com/pkg/errors"

	"myproject/module"
	"myproject/module/utils"
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder: model.DefaultOrder,
				ModulePath:  "myproject/module",
			},
		},
		{
			name: "Imports with inline comments and special aliases",
			input: `package main

import (
	"fmt" // Standard lib
	errlib "github.com/pkg/errors" // Third-party
	utils "myproject/module/utils" // Project package
)
`,
			expected: `package main

import (
	"fmt" // Standard lib

	errlib "github.com/pkg/errors" // Third-party

	utils "myproject/module/utils" // Project package
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder: model.DefaultOrder,
				ModulePath:  "myproject/module",
			},
		},
		{
			name: "Imports with inline comments and special aliases",
			input: `package main

import (
	"fmt" // Standard lib
	errlib "github.com/pkg/errors" // Third-party
	utils "myproject/module/utils" // Project package
)
`,
			expected: `package main

import (
	"fmt"

	errlib "github.com/pkg/errors"

	utils "myproject/module/utils"
)
`,
			wantErr: false,
			opt: &commandline.Option{
				ImportOrder:             model.DefaultOrder,
				ModulePath:              "myproject/module",
				RemoveImportCommentFlag: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := core.FormatImports([]byte(tc.input), tc.opt)
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error status: %v", err)
			}

			if bytes.TrimSpace(output) == nil || bytes.TrimSpace([]byte(tc.expected)) == nil {
				t.Fatalf("unexpected nil output")
			}

			if !bytes.Equal(bytes.TrimSpace(output), bytes.TrimSpace([]byte(tc.expected))) {
				t.Errorf("expected:\n%s\ngot:\n%s", tc.expected, string(output))
			}
		})
	}
}

// formatExact runs FormatImports and compares the whole output.
func formatExact(t *testing.T, input, want string, opt *commandline.Option) {
	t.Helper()
	got, err := core.FormatImports([]byte(input), opt)
	if err != nil {
		t.Fatalf("FormatImports: %v", err)
	}
	if string(got) != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, got)
	}
}

func defaultOpt() *commandline.Option {
	return &commandline.Option{ImportOrder: model.DefaultOrder, ModulePath: "myproject/module"}
}

// The import declaration is found in the syntax tree: an "import (" in a
// comment before it is not the import block.
func TestFormatImports_ImportParenInComment(t *testing.T) {
	formatExact(t, `// Usage: import (
package main

/* see also: import ( "x" ) */
import (
	"os"
	"fmt"
)

func main() { fmt.Println(os.Args) }
`, `// Usage: import (
package main

/* see also: import ( "x" ) */
import (
	"fmt"
	"os"
)

func main() { fmt.Println(os.Args) }
`, defaultOpt())
}

// A parenthesis in a comment inside the import block does not end it.
func TestFormatImports_ParenInImportComment(t *testing.T) {
	formatExact(t, `package main

import (
	"os" // smile :)
	"fmt"
)

func main() { fmt.Println(os.Args) }
`, `package main

import (
	"fmt"
	"os" // smile :)
)

func main() { fmt.Println(os.Args) }
`, defaultOpt())
}

// The same path imported under two names is two imports.
func TestFormatImports_SamePathTwoAliases(t *testing.T) {
	formatExact(t, `package main

import (
	b "fmt"
	"os"
	a "fmt"
)

func main() { a.Println(); b.Println(os.Args) }
`, `package main

import (
	"os"

	a "fmt"
	b "fmt"
)

func main() { a.Println(); b.Println(os.Args) }
`, defaultOpt())
}

// Without a parenthesized import declaration the source is unchanged.
func TestFormatImports_NoImportBlock(t *testing.T) {
	src := "// Usage: import (\npackage main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println() }\n"
	formatExact(t, src, src, defaultOpt())
}

// ReplaceImports fails closed: a result that is not valid Go is an error,
// not output.
func TestReplaceImports_RefusesUnparsableOutput(t *testing.T) {
	src := []byte("package main\n\nimport (\n\t\"fmt\"\n)\n\nfunc main() { fmt.Println() }\n")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	got, err := core.ReplaceImports(src, fset, file, "import (\n\t\"fmt\"\n")
	if err == nil {
		t.Fatalf("expected an error, got output:\n%s", got)
	}
	if got != nil {
		t.Errorf("expected no output with the error, got:\n%s", got)
	}

	got, err = core.ReplaceImports(src, fset, file, "import (\n\t\"fmt\"\n)\n")
	if err != nil || !bytes.Equal(got, src) {
		t.Errorf("valid replacement: got %q, %v", got, err)
	}
}

// --remove-import-comment / -r reaches FormatImports.
func TestFormatImports_RemoveImportCommentFlag(t *testing.T) {
	t.Setenv("GOREG_NOT_USE_CONFIGFILE", "1")
	for _, flag := range []string{"-r", "--remove-import-comment"} {
		t.Run(flag, func(t *testing.T) {
			_, opt, err := commandline.OptParse([]string{flag, "--local", "myproject/module"})
			if err != nil {
				t.Fatal(err)
			}
			formatExact(t, `package main

import (
	// standard library
	"os" // arguments
	"fmt" // printing
)

func main() { fmt.Println(os.Args) }
`, `package main

import (
	"fmt"
	"os"
)

func main() { fmt.Println(os.Args) }
`, opt)
		})
	}
}
