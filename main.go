package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/tools/imports"
)

type importGroup int

const (
	stdLib importGroup = iota
	thirdParty
	local
)

var modulePath string

func getModulePath() string {
	if modulePath != "" {
		return modulePath
	}
	cmd := exec.Command("go", "list", "-m")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("Error: Failed to get module path. Ensure this is a Go module project (with go.mod).")
	}
	modulePath = strings.TrimSpace(string(out))
	return modulePath
}

func formatImports(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	importsMap := make(map[string]struct {
		Doc []string
		End []string
	})
	var stdLibImports, thirdPartyImports, localImports []string

	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		group := getImportGroup(path, modulePath)

		docComments, endComments := extractComments(imp)
		importsMap[path] = struct {
			Doc []string
			End []string
		}{Doc: docComments, End: endComments}

		switch group {
		case stdLib:
			stdLibImports = append(stdLibImports, path)
		case thirdParty:
			thirdPartyImports = append(thirdPartyImports, path)
		case local:
			localImports = append(localImports, path)
		}
	}

	sort.Strings(stdLibImports)
	sort.Strings(thirdPartyImports)
	sort.Strings(localImports)

	var buf bytes.Buffer
	buf.WriteString("import (\n")

	groups := [][]string{}
	if len(stdLibImports) > 0 {
		groups = append(groups, stdLibImports)
	}
	if len(thirdPartyImports) > 0 {
		groups = append(groups, thirdPartyImports)
	}
	if len(localImports) > 0 {
		groups = append(groups, localImports)
	}

	for i, group := range groups {
		writeImports(&buf, group, importsMap, i == 0, i == len(groups)-1)
	}

	buf.WriteString(")\n")

	return replaceImports(src, buf.String()), nil
}

func getImportGroup(pkg string, modulePath string) importGroup {
	if !strings.Contains(pkg, ".") {
		return stdLib
	}
	if strings.HasPrefix(pkg, modulePath) {
		return local
	}
	return thirdParty
}

func writeImports(buf *bytes.Buffer, pkgs []string, importsMap map[string]struct {
	Doc []string
	End []string
}, isFirtGroup bool, isLastGroup bool) {
	for i, imp := range pkgs {
		comments, exists := importsMap[imp]
		if !exists {
			fmt.Fprintf(buf, "\t\"%s\"\n", imp)
			continue
		}

		if !isFirtGroup || i != 0 {
			if len(comments.Doc) > 0 {
				buf.WriteString("\n")
			}
		}

		for _, c := range comments.Doc {
			buf.WriteString(fmt.Sprintf("\t%s\n", c))
		}

		fmt.Fprintf(buf, "\t\"%s\"", imp)

		if len(comments.End) > 0 {
			buf.WriteString(" //")
			for _, c := range comments.End {
				buf.WriteString(fmt.Sprintf(" %s", c))
			}
		}
		buf.WriteString("\n")
	}

	if !isLastGroup {
		buf.WriteString("\n")
	}
}

func extractComments(imp *ast.ImportSpec) ([]string, []string) {
	var docComments []string
	var endComments []string

	if imp.Doc != nil {
		for _, c := range imp.Doc.List {
			docComments = append(docComments, c.Text)
		}
	}
	if imp.Comment != nil {
		for _, c := range imp.Comment.List {
			endComments = append(endComments, c.Text)
		}
	}
	return docComments, endComments
}

func replaceImports(src []byte, newImports string) []byte {
	srcStr := string(src)
	start := strings.Index(srcStr, "import (")
	if start == -1 {
		return src
	}

	end := start
	bracketCount := 1
	for i := start + len("import ("); i < len(srcStr); i++ {
		if srcStr[i] == '(' {
			bracketCount++
		} else if srcStr[i] == ')' {
			bracketCount--
			if bracketCount == 0 {
				end = i + 1
				break
			}
		}
	}

	if end < len(srcStr) && srcStr[end] == '\n' {
		end++
	}

	return []byte(srcStr[:start] + newImports + srcStr[end:])
}

func processFile(filename string, writeToFile bool) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	formatted, err := imports.Process(filename, src, &imports.Options{
		FormatOnly: true,
		Comments:   true,
	})
	if err != nil {
		return err
	}

	sorted, err := formatImports(formatted)
	if err != nil {
		return err
	}

	if writeToFile {
		return os.WriteFile(filename, sorted, 0644)
	} else {
		_, err := os.Stdout.Write(sorted)
		return err
	}
}

func main() {
	flag.Parse()
	writeToFile := flag.Lookup("w") != nil

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: goreg [-w] <file.go>")
		os.Exit(1)
	}
	filename := flag.Arg(0)

	modulePath = getModulePath()

	if err := processFile(filename, writeToFile); err != nil {
		log.Fatal(err)
	}
}
