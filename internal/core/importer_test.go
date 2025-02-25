package core_test

import (
	"bytes"
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
				ImportOrder: model.DefaultOrder,
				ModulePath:  "myproject/module",
				RemoveImportCommentFlag:  true,
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

