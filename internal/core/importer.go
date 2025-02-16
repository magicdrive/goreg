package core

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

func FormatImports(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	importsMap := make(map[string]ImportBlock)
	var stdLibImports, thirdPartyImports, localImports []string

	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		group := GetImportGroup(path, ModulePath)

		docComments, endComments := ExtractComments(imp)
		importsMap[path] = ImportBlock{
			Doc: docComments,
			End: endComments,
		}

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
		WriteImports(&buf, group, importsMap, i == 0, i == len(groups)-1)
	}

	buf.WriteString(")\n")

	return ReplaceImports(src, buf.String()), nil
}

func GetImportGroup(pkg string, modulePath string) ImportGroup {
	if !strings.Contains(pkg, ".") {
		return stdLib
	}
	if strings.HasPrefix(pkg, modulePath) {
		return local
	}
	return thirdParty
}

func WriteImports(buf *bytes.Buffer, pkgs []string, importsMap map[string]ImportBlock, isFirtGroup bool, isLastGroup bool) {
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

func ExtractComments(imp *ast.ImportSpec) ([]string, []string) {
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

func ReplaceImports(src []byte, newImports string) []byte {
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
