// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package linesafe is the FormatOnly path of golang.org/x/tools/imports
// (v0.52.0: internal/imports Process, formatFile, sortImports, sortSpecs,
// mergeImports, the import groups of fix.go, and go/ast/astutil.Imports),
// copied so that one defect can be fixed: those functions computed line
// numbers with token.File.Line, which applies //line directives, and passed
// them to token.File.MergeLine, which takes the line numbers of the file
// itself. With a directive before the import block, x/tools logs "panic
// avoided" forever (sortSpecs) or panics with "invalid line number" (the
// deduplication in sortSpecs). Here every line number is taken with
// PositionFor(pos, false); the changes are marked "goreg:".
//
// goreg uses this package only for sources that contain a line directive
// (HasLineDirective); every other source still goes through
// golang.org/x/tools/imports.Process, whose output this package reproduces
// byte for byte when there is no directive (see the tests). Remove the
// package once x/tools handles line directives.
package linesafe

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// HasLineDirective reports whether src may contain a //line or /*line
// directive. It errs on the side of yes: the text in a string or in another
// comment counts as well, which only sends that source here, where it is
// formatted correctly too.
func HasLineDirective(src []byte) bool {
	return bytes.Contains(src, []byte("//line ")) || bytes.Contains(src, []byte("/*line "))
}

// Process formats src as golang.org/x/tools/imports.Process does with
// &imports.Options{FormatOnly: true, Comments: true} and an empty
// imports.LocalPrefix, the options goreg uses.
func Process(filename string, src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return formatFile(fset, file)
}

// localPrefix is imports.LocalPrefix, which goreg never sets.
const localPrefix = ""

// formatFile is internal/imports.formatFile for a whole file (no fragment,
// so no adjust function) printed with the options above. goreg: the options
// and the printer configuration are constants instead of *Options.
func formatFile(fset *token.FileSet, file *ast.File) ([]byte, error) {
	mergeImports(file)
	sortImports(localPrefix, fset.File(file.FileStart), file)
	var spacesBefore []string // import paths we need spaces before
	for _, impSection := range importSections(fset, file) {
		// Within each block of contiguous imports, see if any
		// import lines are in different group numbers. If so,
		// we'll need to put a space between them so it's
		// compatible with gofmt.
		lastGroup := -1
		for _, importSpec := range impSection {
			importPath, _ := strconv.Unquote(importSpec.Path.Value)
			groupNum := importGroup(localPrefix, importPath)
			if groupNum != lastGroup && lastGroup != -1 {
				spacesBefore = append(spacesBefore, importPath)
			}
			lastGroup = groupNum
		}
	}

	// Comments: true, TabIndent: false, TabWidth: 0.
	printConfig := &printer.Config{Mode: printer.UseSpaces, Tabwidth: 0}

	var buf bytes.Buffer
	err := printConfig.Fprint(&buf, fset, file)
	if err != nil {
		return nil, err
	}
	out := buf.Bytes()
	if len(spacesBefore) > 0 {
		out, err = addImportSpaces(bytes.NewReader(out), spacesBefore)
		if err != nil {
			return nil, err
		}
	}

	out, err = format.Source(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// importSections is go/ast/astutil.Imports: the runs of import specs on
// successive lines.
func importSections(fset *token.FileSet, f *ast.File) [][]*ast.ImportSpec {
	var groups [][]*ast.ImportSpec

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			break
		}

		group := []*ast.ImportSpec{}

		var lastLine int
		for _, spec := range genDecl.Specs {
			importSpec := spec.(*ast.ImportSpec)
			pos := importSpec.Path.ValuePos
			line := fset.PositionFor(pos, false).Line // goreg: was fset.Position(pos).Line
			if lastLine > 0 && pos > 0 && line-lastLine > 1 {
				groups = append(groups, group)
				group = []*ast.ImportSpec{}
			}
			group = append(group, importSpec)
			lastLine = line
		}
		groups = append(groups, group)
	}

	return groups
}

var importToGroup = []func(localPrefix, importPath string) (num int, ok bool){
	func(localPrefix, importPath string) (num int, ok bool) {
		if localPrefix == "" {
			return
		}
		for p := range strings.SplitSeq(localPrefix, ",") {
			if strings.HasPrefix(importPath, p) || strings.TrimSuffix(p, "/") == importPath {
				return 3, true
			}
		}
		return
	},
	func(_, importPath string) (num int, ok bool) {
		if strings.HasPrefix(importPath, "appengine") {
			return 2, true
		}
		return
	},
	func(_, importPath string) (num int, ok bool) {
		firstComponent := strings.Split(importPath, "/")[0]
		if strings.Contains(firstComponent, ".") {
			return 1, true
		}
		return
	},
}

func importGroup(localPrefix, importPath string) int {
	for _, fn := range importToGroup {
		if n, ok := fn(localPrefix, importPath); ok {
			return n
		}
	}
	return 0
}

var impLine = regexp.MustCompile(`^\s+(?:[\w\.]+\s+)?"(.+?)"`)

func addImportSpaces(r io.Reader, breaks []string) ([]byte, error) {
	var out bytes.Buffer
	in := bufio.NewReader(r)
	inImports := false
	done := false
	for {
		s, err := in.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if !inImports && !done && strings.HasPrefix(s, "import") {
			inImports = true
		}
		if inImports && (strings.HasPrefix(s, "var") ||
			strings.HasPrefix(s, "func") ||
			strings.HasPrefix(s, "const") ||
			strings.HasPrefix(s, "type")) {
			done = true
			inImports = false
		}
		if inImports && len(breaks) > 0 {
			if m := impLine.FindStringSubmatch(s); m != nil {
				if m[1] == breaks[0] {
					out.WriteByte('\n')
					breaks = breaks[1:]
				}
			}
		}

		fmt.Fprint(&out, s)
	}
	return out.Bytes(), nil
}
