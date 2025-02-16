package core_test

import (
	"bytes"
	"testing"

	"github.com/magicdrive/goreg/internal/core"
)

func TestFormatImports(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		expected string
		wantErr bool
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

