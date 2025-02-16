package core_test

import (
	"bytes"
	"testing"

	"github.com/magicdrive/goreg/internal/core"
)

func TestFormatImports(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
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
		},
		{
			name: "Aliased imports",
			input: `package main

import (
	cmd "github.com/hogehoge/cmd/tools"
	"fmt"
	"github.com/pkg/errors"
)
`,
			expected: `package main

import (
	"fmt"

	"github.com/pkg/errors"

	cmd "github.com/hogehoge/cmd/tools"
)
`,
			wantErr: false,
		},
		{
			name: "Multiple aliased imports",
			input: `package main

import (
	cmd "github.com/hogehoge/cmd/tools"
	log "github.com/sirupsen/logrus"
	"fmt"
	"github.com/pkg/errors"
)
`,
			expected: `package main

import (
	"fmt"

	"github.com/pkg/errors"

	cmd "github.com/hogehoge/cmd/tools"
	log "github.com/sirupsen/logrus"
)
`,
			wantErr: false,
		},
		{
			name: "Imports with line comments",
			input: `package main

import (
	"fmt" // Standard library
	"github.com/pkg/errors" // Third-party package
	"myproject/module" // Project package
)
`,
			expected: `package main

import (
	"fmt" // Standard library

	"github.com/pkg/errors" // Third-party package

	"myproject/module" // Project package
)
`,
			wantErr: false,
		},
		{
			name: "Aliased imports with comments",
			input: `package main

import (
	cmd "github.com/hogehoge/cmd/tools" // Command tools
	"fmt" // Standard library
	"github.com/pkg/errors" // Error handling
)
`,
			expected: `package main

import (
	"fmt" // Standard library

	"github.com/pkg/errors" // Error handling

	cmd "github.com/hogehoge/cmd/tools" // Command tools
)
`,
			wantErr: false,
		},
		{
			name: "Imports with block comments",
			input: `package main

import (
	/* Standard library */
	"fmt"

	/* Error handling */
	"github.com/pkg/errors"

	/* Project-specific module */
	"myproject/module"
)
`,
			expected: `package main

import (
	/* Standard library */
	"fmt"

	/* Error handling */
	"github.com/pkg/errors"

	/* Project-specific module */
	"myproject/module"
)
`,
			wantErr: false,
		},
		{
			name: "Mixed line and block comments",
			input: `package main

import (
	/* Standard lib */ "fmt"
	"github.com/pkg/errors" // Error handling
	"myproject/module" /* Project module */
)
`,
			expected: `package main

import (
	"fmt"

	"github.com/pkg/errors" // Error handling

	"myproject/module" /* Project module */
)
`,
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := core.FormatImports([]byte(tc.input), "myproject/module")
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

